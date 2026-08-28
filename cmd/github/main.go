package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/hex"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/csmith/envflag/v2"
	"github.com/csmith/slogflags"
	"github.com/greboid/irc-bot/v5/plugins"
	"github.com/greboid/irc-bot/v5/rpc"
)

var (
	RPCHost        = flag.String("rpc-host", "localhost", "gRPC server to connect to")
	RPCPort        = flag.Int("rpc-port", 8001, "gRPC server port")
	RPCToken       = flag.String("rpc-token", "", "gRPC authentication token")
	Channel        = flag.String("channel", "", "Channel to send messages to")
	PrivateChannel = flag.String("private-channel", "", "Channel to send messages to")
	HidePrivate    = flag.Bool("hide-private", false, "Hide notifications about private repos")
	GithubSecret   = flag.String("github-secret", "", "Github secret for validating webhooks")
	IgnoredUsers   = flag.String("ignored-users", "", "Comma separated list of users to ignore")
	helper         *plugins.PluginHelper
)

func main() {
	envflag.Parse()
	slogflags.Logger(slogflags.WithSetDefault(true))
	slog.Info("Starting github plugin")
	var err error
	helper, err = plugins.NewHelper(fmt.Sprintf("%s:%d", *RPCHost, uint16(*RPCPort)), *RPCToken)
	if err != nil {
		slog.Error("Unable to create plugin helper: %s", err.Error())
		os.Exit(1)
		return
	}
	err = helper.RegisterWebhook("github", handleGithub)
	if err != nil {
		slog.Error("Error registering webhook: %s", err.Error())
		os.Exit(1)
		return
	}
	slog.Info("Exiting")
}

func handleGithub(request *rpc.HttpRequest) *rpc.HttpResponse {
	headers := rpc.ConvertFromRPCHeaders(request.Header)
	eventType := headers.Get("X-GitHub-Event")
	header := strings.SplitN(headers.Get("X-Hub-Signature"), "=", 2)
	if header[0] != "sha1" {
		slog.Debug("Error: %s", "Bad header")
		return &rpc.HttpResponse{
			Header: nil,
			Body:   []byte("Bad headers"),
			Status: http.StatusInternalServerError,
		}
	}
	if !CheckGithubSecret(request.Body, header[1], *GithubSecret) {
		slog.Debug("Error: %s", "Bad hash")
		return &rpc.HttpResponse{
			Header: nil,
			Body:   []byte("Bad hash"),
			Status: http.StatusBadRequest,
		}
	}
	go func() {
		slog.Info("Received github notification: %s", eventType)
		webhookHandler := githubWebhookHandler{
			sender:         helper,
			ignoredSenders: parseIgnoredUsers(*IgnoredUsers),
		}
		err := webhookHandler.handleWebhook(eventType, request.Body)
		if err != nil {
			slog.Error("Unable to handle webhook: %s", err.Error())
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

func parseIgnoredUsers(users string) []string {
	ignoredUsers := make([]string, 0)
	splitUsers := strings.Split(users, ",")
	for _, user := range splitUsers {
		trimmedUser := strings.TrimSpace(user)
		if trimmedUser != "" {
			ignoredUsers = append(ignoredUsers, trimmedUser)
		}
	}
	return ignoredUsers
}
