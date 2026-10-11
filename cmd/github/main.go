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
	ForgejoSecret  = flag.String("forgejo-secret", "", "Forgejo secret for validating webhooks")
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
		slog.Error("Unable to create plugin helper", slog.Any("error", err))
		os.Exit(1)
		return
	}
	// RegisterWebhook blocks for the lifetime of the stream, so each webhook
	// needs its own goroutine to get both registered.
	errs := make(chan error, 2)
	go func() { errs <- helper.RegisterWebhook("github", handleGithub) }()
	go func() { errs <- helper.RegisterWebhook("forgejo", handleForgejo) }()
	for range 2 {
		if err := <-errs; err != nil {
			slog.Error("Error registering webhook", slog.Any("error", err))
			os.Exit(1)
		}
	}
	slog.Info("Exiting")
}

func handleGithub(request *rpc.HttpRequest) *rpc.HttpResponse {
	headers := rpc.ConvertFromRPCHeaders(request.Header)
	eventType := headers.Get("X-GitHub-Event")
	signature, ok := parseSignatureHeader(headers.Get("X-Hub-Signature"))
	if !ok {
		slog.Debug("Bad header")
		return &rpc.HttpResponse{
			Header: nil,
			Body:   []byte("Bad headers"),
			Status: http.StatusInternalServerError,
		}
	}
	if !CheckGithubSecret(request.Body, signature, *GithubSecret) {
		slog.Debug("Bad hash")
		return &rpc.HttpResponse{
			Header: nil,
			Body:   []byte("Bad hash"),
			Status: http.StatusBadRequest,
		}
	}
	go func() {
		slog.Info("Received github notification", slog.Any("event", eventType))
		webhookHandler := githubWebhookHandler{
			sender:         helper,
			ignoredSenders: parseIgnoredUsers(*IgnoredUsers),
		}
		err := webhookHandler.handleWebhook(eventType, request.Body)
		if err != nil {
			slog.Error("Unable to handle webhook", slog.Any("error", err))
		}
	}()
	return &rpc.HttpResponse{
		Header: nil,
		Body:   []byte("Delivered"),
		Status: http.StatusOK,
	}
}

func parseSignatureHeader(header string) (string, bool) {
	parts := strings.SplitN(header, "=", 2)
	if len(parts) != 2 || parts[0] != "sha1" {
		return "", false
	}
	return parts[1], true
}

func CheckGithubSecret(bodyBytes []byte, headerSecret string, githubSecret string) bool {
	h := hmac.New(sha1.New, []byte(githubSecret))
	h.Write(bodyBytes)
	expected := hex.EncodeToString(h.Sum(nil))
	return len(expected) == len(headerSecret) && subtle.ConstantTimeCompare([]byte(expected), []byte(headerSecret)) == 1
}

func parseIgnoredUsers(users string) []string {
	ignoredUsers := make([]string, 0)
	splitUsers := strings.SplitSeq(users, ",")
	for user := range splitUsers {
		trimmedUser := strings.TrimSpace(user)
		if trimmedUser != "" {
			ignoredUsers = append(ignoredUsers, trimmedUser)
		}
	}
	return ignoredUsers
}
