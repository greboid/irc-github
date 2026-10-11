package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/greboid/irc-bot/v5/rpc"
)

func handleForgejo(request *rpc.HttpRequest) *rpc.HttpResponse {
	headers := rpc.ConvertFromRPCHeaders(request.Header)
	eventType := forgejoEventType(headers)
	if !CheckForgejoSecret(request.Body, forgejoSignature(headers), *ForgejoSecret) {
		slog.Debug("Bad forgejo signature")
		return &rpc.HttpResponse{
			Header: nil,
			Body:   []byte("Bad signature"),
			Status: http.StatusBadRequest,
		}
	}
	go func() {
		slog.Info("Received forgejo notification", slog.Any("event", eventType))
		webhookHandler := &forgejoWebhookHandler{
			gh: &githubWebhookHandler{
				sender:         helper,
				ignoredSenders: parseIgnoredUsers(*IgnoredUsers),
			},
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

// forgejoEventType prefers forgejo headers and falls back to the gitea ones
// sent by forgejo for backwards compatibility.
func forgejoEventType(headers http.Header) string {
	if event := headers.Get("X-Forgejo-Event"); event != "" {
		return event
	}
	return headers.Get("X-Gitea-Event")
}

func forgejoSignature(headers http.Header) string {
	if signature := headers.Get("X-Forgejo-Signature"); signature != "" {
		return signature
	}
	return headers.Get("X-Gitea-Signature")
}

func CheckForgejoSecret(bodyBytes []byte, headerSecret string, forgejoSecret string) bool {
	h := hmac.New(sha256.New, []byte(forgejoSecret))
	h.Write(bodyBytes)
	expected := hex.EncodeToString(h.Sum(nil))
	return len(expected) == len(headerSecret) && subtle.ConstantTimeCompare([]byte(expected), []byte(headerSecret)) == 1
}

// Forgejo payloads are close enough to github's that the github handlers are
// reused; the github.go helpers cover the field differences.
var forgejoEventHandlers = map[string]func(webhook) []string{
	"push":          (&githubPushHandler{}).handlePushEvent,
	"pull_request":  (&githubPRHandler{}).handlePREvent,
	"issues":        (&githubissuehandler{}).handleIssueEvent,
	"issue_comment": (&githubIssueCommenthandler{}).handleIssueCommentEvent,
	"release":       (&githubReleaseHandler{}).handleReleaseEvent,
	"create":        (&githubCreateHandler{}).handleCreateEvent,
	"delete":        handleForgejoDeleteEvent,
	"package":       (&githubPackageHandler{}).handlePackageEvent,
}

func handleForgejoDeleteEvent(data webhook) (messages []string) {
	messages = append(messages, fmt.Sprintf(
		"[%s] %s deleted %s %s",
		data.Repository.FullName,
		data.Sender.Login,
		data.RefType,
		data.Refspec,
	))
	return
}

type forgejoWebhookHandler struct {
	gh *githubWebhookHandler
}

func (f *forgejoWebhookHandler) handleWebhook(eventType string, bodyBytes []byte) error {
	if eventType == "ping" {
		// Forgejo ping payloads contain no repository or user data.
		return nil
	}
	handler, ok := forgejoEventHandlers[eventType]
	if !ok {
		return nil
	}
	data := webhook{}
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		return err
	}
	f.gh.hookSender = data.Sender.Login
	if err := f.gh.sendMessage(data.Repository.IsPrivate, handler(data)...); err != nil {
		slog.Error("Error handling forgejo event", slog.Any("event", eventType), slog.Any("error", err))
	}
	return nil
}
