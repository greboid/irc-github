package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/greboid/irc-bot/v4/plugins"
	"github.com/greboid/irc-bot/v4/rpc"
	"github.com/greboid/irc/v4/logger"
	"github.com/kouhin/envflag"
	"go.uber.org/zap"
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
	log            = logger.CreateLogger(*Debug)
	helper         *plugins.PluginHelper
)

type github struct {
	client rpc.IRCPluginClient
	log    *zap.SugaredLogger
}

func main() {
	log.Infof("Starting github plugin")
	err := envflag.Parse()
	if err != nil {
		log.Fatalf("Unable to load config: %s", err.Error())
		return
	}
	helper, err = plugins.NewHelper(fmt.Sprintf("%s:%d", *RPCHost, uint16(*RPCPort)), *RPCToken)
	if err != nil {
		log.Fatalf("Unable to create plugin helper: %s", err.Error())
		return
	}
	err = helper.RegisterWebhook("github", handleGithub)
	if err != nil {
		log.Fatalf("Error registering webhook: %s", err.Error())
		return
	}
	log.Infof("Exiting")
}

func handleGithub(request *rpc.HttpRequest) *rpc.HttpResponse {
	g := github{
		log: log,
	}
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
		log.Infof("Received github notification: %s", eventType)
		webhookHandler := githubWebhookHandler{}
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
