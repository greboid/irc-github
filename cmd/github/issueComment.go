package main

import (
	"fmt"
)

type githubIssueCommenthandler struct{}

func (g *githubIssueCommenthandler) handleIssueCommentEvent(data issuehook) (messages []string) {
	switch data.Action {
	case "created":
		return g.handleIssueCommentCreated(data)
	}
	return []string{}
}

func (g *githubIssueCommenthandler) handleIssueCommentCreated(data issuehook) (messages []string) {
	messages = append(messages, fmt.Sprintf(
		"[%s] %s commented on issue %s - %s",
		data.Repository.FullName,
		data.User.Login,
		data.Issue.Title,
		data.Issue.HtmlURL,
	))
	return
}
