package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/subtle"
	"crypto/tls"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/greboid/irc/v2/logger"
	"github.com/greboid/irc/v2/rpc"
	"github.com/kouhin/envflag"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io"
	"net/http"
	"strings"
)

var (
	RPCHost        = flag.String("rpc-host", "localhost", "gRPC server to connect to")
	RPCPort        = flag.Int("rpc-port", 8001, "gRPC server port")
	RPCToken       = flag.String("rpc-token", "", "gRPC authentication token")
	Channel        = flag.String("channel", "", "Channel to send messages to")
	PrivateChannel = flag.String("private-channel", "", "Channel to send messages to")
	HidePrivate    = flag.Bool("hide-private", false, "Hide notifications about private repos")
	GithubSecret   = flag.String("github-secret", "", "Github secret for validating webhooks")
	Debug          = flag.Bool("debug", false, "Show debugging info")
)

type github struct {
	client rpc.IRCPluginClient
	log    *zap.SugaredLogger
}

func main() {
	log := logger.CreateLogger(*Debug)
	if err := envflag.Parse(); err != nil {
		log.Fatalf("Unable to load config: %s", err.Error())
	}
	github := github{
		log: log,
	}
	log.Infof("Creating Github RPC Client")
	client, err := github.doRPC()
	if err != nil {
		log.Fatalf("Unable to create RPC Client: %s", err.Error())
	}
	github.client = client
	log.Infof("Starting github web server")
	err = github.doWeb()
	if err != nil {
		log.Panicf("Error handling web: %s", err.Error())
	}
	log.Infof("exiting")
}

func (g *github) doRPC() (rpc.IRCPluginClient, error) {
	creds := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", *RPCHost, *RPCPort), grpc.WithTransportCredentials(creds))
	client := rpc.NewIRCPluginClient(conn)
	_, err = client.Ping(rpc.CtxWithToken(context.Background(), "bearer", *RPCToken), &rpc.Empty{})
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (g *github) doWeb() error {
	creds := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", *RPCHost, *RPCPort), grpc.WithTransportCredentials(creds))
	if err != nil {
		return err
	}
	client := rpc.NewHTTPPluginClient(conn)
	stream, err := client.GetRequest(rpc.CtxWithTokenAndPath(context.Background(), "bearer", *RPCToken, "github"))
	for {
		request, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return nil
		}
		response := g.handleGithub(request)
		err = stream.Send(response)
		if err != nil {
			return err
		}
	}
}

func (g *github) handleGithub(request *rpc.HttpRequest) *rpc.HttpResponse {
	headers := rpc.ConvertFromRPCHeaders(request.Header)
	eventType := headers.Get("X-GitHub-Event")
	header := strings.SplitN(headers.Get("X-Hub-Signature"), "=", 2)
	if header[0] != "sha1" {
		g.log.Debugf("Error: %s", "Bad header")
		return &rpc.HttpResponse{
			Header: nil,
			Body:   []byte("Bad headers"),
			Status: http.StatusInternalServerError,
		}
	}
	if !CheckGithubSecret(request.Body, header[1], *GithubSecret) {
		g.log.Debugf("Error: %s", "Bad hash")
		return &rpc.HttpResponse{
			Header: nil,
			Body:   []byte("Bad hash"),
			Status: http.StatusBadRequest,
		}
	}
	go func() {
		webhookHandler := githubWebhookHandler{
			client: g.client,
		}
		err := webhookHandler.handleWebhook(eventType, request.Body)
		if err != nil {
			g.log.Errorf("Unable to handle webhook: %s", err.Error())
		}
	}()
	return &rpc.HttpResponse{
		Header: nil,
		Body:   []byte("Delivered"),
		Status: http.StatusOK,
	}
}

func CheckGithubSecret(bodyBytes []byte, headerSecret string, githubSecret string) bool {
	h := hmac.New(sha1.New, []byte(githubSecret))
	h.Write(bodyBytes)
	expected := fmt.Sprintf("%s", hex.EncodeToString(h.Sum(nil)))
	return len(expected) == len(headerSecret) && subtle.ConstantTimeCompare([]byte(expected), []byte(headerSecret)) == 1
}
