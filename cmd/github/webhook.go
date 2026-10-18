package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strings"
)

type messageSender interface {
	SendChannelMessage(channel string, messages ...string) error
}

type githubWebhookHandler struct {
	sender         messageSender
	hookSender     string
	ignoredSenders []string
	rewriteRules   []rewriteRule
}

func (g *githubWebhookHandler) handleWebhook(eventType string, bodyBytes []byte) error {
	switch eventType {
	case "ping":
		data := webhook{}
		err := json.Unmarshal(bodyBytes, &data)
		if err == nil {
			go func() {
				g.hookSender = data.Sender.Login
				err := g.sendMessage(data.Repository.IsPrivate, fmt.Sprintf("Ping received for %s", data.Repository.FullName))
				if err != nil {
					slog.Error("Error handling ping", slog.Any("error", err))
				}
			}()
		} else {
			slog.Error("Error handling ping", slog.Any("error", err))
			return err
		}
	case "push":
		data := webhook{}
		handler := githubPushHandler{}
		err := json.Unmarshal(bodyBytes, &data)
		if err == nil {
			go func() {
				g.hookSender = data.Sender.Login
				err := g.sendMessage(data.Repository.IsPrivate, handler.handlePushEvent(data)...)
				if err != nil {
					slog.Error("Error handling push", slog.Any("error", err))
				}
			}()
		} else {
			slog.Error("Error handling push", slog.Any("error", err))
			return err
		}
	case "pull_request":
		data := webhook{}
		handler := githubPRHandler{}
		err := json.Unmarshal(bodyBytes, &data)
		if err == nil {
			go func() {
				g.hookSender = data.Sender.Login
				err := g.sendMessage(data.Repository.IsPrivate, handler.handlePREvent(data)...)
				if err != nil {
					slog.Error("Error handling pull request", slog.Any("error", err))
				}
			}()
		} else {
			slog.Error("Error handling pull request", slog.Any("error", err))
			return err
		}
	case "issues":
		data := webhook{}
		handler := githubissuehandler{}
		err := json.Unmarshal(bodyBytes, &data)
		if err == nil {
			go func() {
				g.hookSender = data.Sender.Login
				err := g.sendMessage(data.Repository.IsPrivate, handler.handleIssueEvent(data)...)
				if err != nil {
					slog.Error("Error handling issues", slog.Any("error", err))
				}
			}()
		} else {
			slog.Error("Error handling issues", slog.Any("error", err))
			return err
		}
	case "issue_comment":
		data := webhook{}
		handler := githubIssueCommenthandler{}
		err := json.Unmarshal(bodyBytes, &data)
		if err == nil {
			go func() {
				g.hookSender = data.Sender.Login
				err := g.sendMessage(data.Repository.IsPrivate, handler.handleIssueCommentEvent(data)...)
				if err != nil {
					slog.Error("Error handling issue comment", slog.Any("error", err))
				}
			}()
		} else {
			slog.Error("Error handling issue comment", slog.Any("error", err))
			return err
		}
	case "check_run":
		data := webhook{}
		handler := githubCheckRunHandler{}
		err := json.Unmarshal(bodyBytes, &data)
		if err == nil {
			go func() {
				g.hookSender = data.Sender.Login
				err := g.sendMessage(data.Repository.IsPrivate, handler.handleCheckRunEvent(data)...)
				if err != nil {
					slog.Error("Error handling check run", slog.Any("error", err))
				}
			}()
		} else {
			slog.Error("Error handling check run", slog.Any("error", err))
			return err
		}
	case "release":
		data := webhook{}
		handler := githubReleaseHandler{}
		err := json.Unmarshal(bodyBytes, &data)
		if err == nil {
			go func() {
				g.hookSender = data.Sender.Login
				err := g.sendMessage(data.Repository.IsPrivate, handler.handleReleaseEvent(data)...)
				if err != nil {
					slog.Error("Error handling release", slog.Any("error", err))
				}
			}()
		} else {
			slog.Error("Error handling release", slog.Any("error", err))
			return err
		}
	case "create":
		data := webhook{}
		handler := githubCreateHandler{}
		err := json.Unmarshal(bodyBytes, &data)
		if err == nil {
			go func() {
				g.hookSender = data.Sender.Login
				err := g.sendMessage(data.Repository.IsPrivate, handler.handleCreateEvent(data)...)
				if err != nil {
					slog.Error("Error handling create", slog.Any("error", err))
				}
			}()
		} else {
			slog.Error("Error handling create", slog.Any("error", err))
			return err
		}
	case "package":
		data := webhook{}
		handler := githubPackageHandler{}
		err := json.Unmarshal(bodyBytes, &data)
		if err == nil {
			go func() {
				g.hookSender = data.Sender.Login
				err := g.sendMessage(data.Repository.IsPrivate, handler.handlePackageEvent(data)...)
				if err != nil {
					slog.Error("Error handling package", slog.Any("error", err))
				}
			}()
		} else {
			slog.Error("Error handling package", slog.Any("error", err))
			return err
		}
	case "check_suite":
		data := webhook{}
		handler := githubCheckSuiteHandler{}
		err := json.Unmarshal(bodyBytes, &data)
		if err == nil {
			go func() {
				g.hookSender = data.Sender.Login
				err := g.sendMessage(data.Repository.IsPrivate, handler.handleCheckSuiteEvent(data)...)
				if err != nil {
					slog.Error("Error handling check suite", slog.Any("error", err))
				}
			}()
		} else {
			slog.Error("Error handling check suite", slog.Any("error", err))
			return err
		}
	}
	return nil
}

func (g *githubWebhookHandler) sendMessage(isPrivate bool, messages ...string) error {
	sender := g.hookSender
	if slices.Contains(g.ignoredSenders, sender) {
		return nil
	}
	notifyChannel := *Channel
	if isPrivate && *HidePrivate {
		return nil
	}
	if isPrivate && len(*PrivateChannel) != 0 {
		notifyChannel = *PrivateChannel
	}
	for i := range messages {
		messages[i] = rewriteMessage(g.rewriteRules, messages[i])
	}
	return g.sender.SendChannelMessage(notifyChannel, messages...)
}

func rewriteMessage(rules []rewriteRule, message string) string {
	for _, rule := range rules {
		message = strings.ReplaceAll(message, rule.from, rule.to)
	}
	return message
}
