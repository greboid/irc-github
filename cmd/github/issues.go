package main

import "fmt"

type githubissuehandler struct{}

func (g *githubissuehandler) handleIssueEvent(data webhook) (messages []string) {
	switch data.Action {
	case "opened":
		return g.handleIssueOpened(data)
	case "closed":
		return g.handleIssueClosed(data)
	}
	return []string{}
}

func (g *githubissuehandler) handleIssueOpened(data webhook) (messages []string) {
	messages = append(messages, fmt.Sprintf(
		"[%s] %s create issue: %s -  %s",
		data.Repository.FullName,
		data.Sender.Login,
		data.Issue.Title,
		data.Issue.HtmlURL,
	))
	return
}

func (g *githubissuehandler) handleIssueClosed(data webhook) (messages []string) {
	messages = append(messages, fmt.Sprintf(
		"[%s] %s closed issue %s - %s",
		data.Repository.FullName,
		data.Sender.Login,
		data.Issue.Title,
		data.Issue.HtmlURL,
	))
	return
}
