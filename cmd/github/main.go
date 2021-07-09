package main

import (
	"flag"
	"fmt"
	webhook "github.com/go-playground/webhooks/v6/github"
	"github.com/greboid/irc-bot/v4/plugins"
	"github.com/greboid/irc-bot/v4/rpc"
	"github.com/greboid/irc/v4/logger"
	"github.com/kouhin/envflag"
	"go.uber.org/zap"
	"net/http"
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
	hook, err := webhook.New(webhook.Options.Secret(*GithubSecret))
	if err != nil {
		g.log.Debugf("Error: %s", "not able to create webhook")
		return getError(http.StatusInternalServerError, "not able to create webhook: " + err.Error())
	}
	payload, err := hook.Parse(rpc.ConvertRPCToHTTP(request),
		webhook.PingEvent,
		webhook.IssuesEvent, webhook.IssueCommentEvent,
		webhook.PullRequestEvent,
		webhook.PushEvent,
	)
	if err != nil {
		if err == webhook.ErrEventNotFound {
			return getError(http.StatusUnprocessableEntity, "Unknown event type")
		}
		return getError(http.StatusInternalServerError, "not able to create webhook: " + err.Error())
	}
	gh := githubWebhookHandler{
		sender: helper,
	}
	private, messages := gh.handleWebhook(payload)
	err = gh.sendMessage(private, messages...)
	if err != nil {
		g.log.Errorf("Unable to send messages: %s", err.Error())
		return &rpc.HttpResponse{
			Header: nil,
			Body:   []byte("Error delivering message"),
			Status: http.StatusInternalServerError,
		}
	} else {
		return &rpc.HttpResponse{
			Header: nil,
			Body:   []byte("Delivered"),
			Status: http.StatusOK,
		}
	}
}