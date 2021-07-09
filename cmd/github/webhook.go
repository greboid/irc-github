package main

import (
	"fmt"
	webhook "github.com/go-playground/webhooks/v6/github"
	"github.com/greboid/irc-bot/v4/rpc"
	"strings"
)

type messageSender interface {
	SendChannelMessage(channel string, messages... string) error
}

type githubWebhookHandler struct {
	sender messageSender
}

func (gh *githubWebhookHandler) handleWebhook(payload interface{}) (bool, []string) {
	switch payload.(type) {
	case webhook.PingPayload:
		return gh.handlePing(payload.(webhook.PingPayload))
	case webhook.IssuesPayload:
		return gh.handleIssues(payload.(webhook.IssuesPayload))
	case webhook.IssueCommentPayload:
		return gh.handleIssueComments(payload.(webhook.IssueCommentPayload))
	case webhook.PullRequestPayload:
		return gh.handlePullRequest(payload.(webhook.PullRequestPayload))
	case webhook.PushPayload:
		return gh.handlePush(payload.(webhook.PushPayload))
	default:
		log.Infof("Unable to handle webhook, unknown type")
	}
	return false, nil
}

func (gh *githubWebhookHandler) handlePush(data webhook.PushPayload) (bool, []string) {
	if data.Created {
		return handlePushCreate(data)
	} else if data.Deleted {
		return handlePushDelete(data)
	} else {
		return handlePushCommit(data)
	}
}

func handlePushCreate(data webhook.PushPayload) (isPrivate bool, messages []string) {
	isPrivate = data.Repository.Private
	if *data.BaseRef == "" {
		messages = append(messages, fmt.Sprintf(
			"[%s] %s created %s - %s",
			data.Repository.FullName, data.Pusher.Name, tidyRef(data.Ref), data.Compare,
		))
	} else {
		messages = append(messages, fmt.Sprintf(
			"[%s] %s created %s from %s - %s",
			data.Repository.FullName, data.Pusher.Name, tidyRef(data.Ref), tidyBaseRef(data.BaseRef), data.Compare,
		))
	}
	return
}

func handlePushDelete(data webhook.PushPayload) (isPrivate bool, messages []string) {
	isPrivate = data.Repository.Private
	messages = append(messages, fmt.Sprintf(
		"[%s] %s deleted %s",
		data.Repository.FullName, data.Sender.Login, tidyRef(data.Ref),
	))
	return
}

func handlePushCommit(data webhook.PushPayload) (isPrivate bool, messages []string) {
	isPrivate = data.Repository.Private
	messages = append(messages, fmt.Sprintf(
		"[%s] %s pushed %d commits to %s - %s",
		data.Repository.FullName, data.Pusher.Name, len(data.Commits), tidyRef(data.Ref), data.Compare,
	))
	for _, commit := range data.Commits {
		messages = append(messages, fmt.Sprintf(
			"[%s] %s committed %s - %s",
			data.Repository.FullName, commit.Author.Username, commit.ID[len(commit.ID)-6:], strings.SplitN(commit.Message, "\n", 2)[0],
		))
	}
	return
}

func (gh *githubWebhookHandler) handlePullRequest(data webhook.PullRequestPayload) (isPrivate bool, messages []string) {
	isPrivate = data.Repository.Private
	switch data.Action {
	case "opened":
		messages = append(messages, fmt.Sprintf(
			"[%s] %s submitted PR: %s -  %s",
			data.Repository.FullName, data.PullRequest.User.Login, data.PullRequest.Title, data.PullRequest.HTMLURL,
		))
	case "closed":
		if data.PullRequest.Merged {
			messages = append(messages, fmt.Sprintf(
				"[%s] %s closed PR: %s -  %s",
				data.Repository.FullName, data.PullRequest.User.Login, data.PullRequest.Title, data.PullRequest.HTMLURL,
			))
		} else {
			messages = append(messages, fmt.Sprintf(
				"[%s] %s merged PR from %s: %s -  %s",
				data.Repository.FullName, data.PullRequest.MergedBy.Login, data.PullRequest.User.Login, data.PullRequest.Title, data.PullRequest.HTMLURL,
			))
		}
	}
	return
}

func (gh *githubWebhookHandler) handleIssueComments(data webhook.IssueCommentPayload) (isPrivate bool, messages []string) {
	isPrivate = data.Repository.Private
	switch data.Action {
	case "opened":
		messages = append(messages, fmt.Sprintf(
			"[%s] %s commented on issue %s - %s",
			data.Repository.FullName, data.Sender.Login, data.Issue.Title, data.Issue.HTMLURL,
		))
	}
	return
}

func (gh *githubWebhookHandler)handleIssues(data webhook.IssuesPayload) (isPrivate bool, messages []string) {
	isPrivate = data.Repository.Private
	switch data.Action {
	case "opened":
		messages = append(messages, fmt.Sprintf(
			"[%s] %s create issue: %s -  %s",
			data.Repository.FullName, data.Sender.Login, data.Issue.Title, data.Issue.HTMLURL,
		))
	case "closed":
		messages = append(messages, fmt.Sprintf(
			"[%s] %s closed issue %s - %s",
			data.Repository.FullName, data.Sender.Login, data.Issue.Title, data.Issue.HTMLURL,
		))
	}
	return
}

func (gh *githubWebhookHandler) handlePing(data webhook.PingPayload) (isPrivate bool, messages []string) {
	isPrivate = data.Repository.Private
	messages = append(messages, fmt.Sprintf("Ping received for %s", data.Repository.FullName))
	return
}

func getError(status int32, message string) *rpc.HttpResponse {
	return &rpc.HttpResponse{ Header: nil, Body:   []byte(message), Status: status }
}

func tidyRef(data string) string {
	if strings.HasPrefix(data, "refs/heads/") {
		return fmt.Sprintf("branch %s", strings.TrimPrefix(data, "refs/heads/"))
	}
	if strings.HasPrefix(data, "refs/tags/") {
		return fmt.Sprintf("tag %s", strings.TrimPrefix(data, "refs/tags/"))
	}
	return data
}

func tidyBaseRef(data *string) string {
	if strings.HasPrefix(*data, "refs/heads/") {
		return fmt.Sprintf("branch %s", strings.TrimPrefix(*data, "refs/heads/"))
	}
	if strings.HasPrefix(*data, "refs/tags/") {
		return fmt.Sprintf("tag %s", strings.TrimPrefix(*data, "refs/tags/"))
	}
	return *data
}

func (gh *githubWebhookHandler) sendMessage(isPrivate bool, messages ...string) error {
	notifyChannel := *Channel
	if isPrivate && *HidePrivate {
		return nil
	}
	if isPrivate && len(*PrivateChannel) != 0 {
		notifyChannel = *PrivateChannel
	}
	return gh.sender.SendChannelMessage(notifyChannel, messages...)
}
