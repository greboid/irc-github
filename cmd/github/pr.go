package main

import "fmt"

type githubPRHandler struct{}

func (g *githubPRHandler) handlePREvent(data prhook) (messages []string) {
	if data.Action == "opened" {
		return g.handlePROpen(data)
	} else if data.Action == "closed" {
		if data.PullRequest.Merged == "" {
			return g.handlePRClose(data)
		} else {
			return g.handlePRMerged(data)
		}
	}
	return
}

func (g *githubPRHandler) handlePRClose(data prhook) (messages []string) {
	messages = append(messages, fmt.Sprintf(
		"[%s] %s closed PR: %s -  %s",
		data.Repository.FullName,
		data.PullRequest.User.Login,
		data.PullRequest.Title,
		data.PullRequest.HtmlURL,
	))
	return
}

func (g *githubPRHandler) handlePRMerged(data prhook) (messages []string) {
	messages = append(messages, fmt.Sprintf(
		"[%s] %s merged PR from %s: %s -  %s",
		data.Repository.FullName,
		data.PullRequest.MergedBy.Login,
		data.PullRequest.User.Login,
		data.PullRequest.Title,
		data.PullRequest.HtmlURL,
	))
	return
}

func (g *githubPRHandler) handlePROpen(data prhook) (messages []string) {
	messages = append(messages, fmt.Sprintf(
		"[%s] %s submitted PR: %s -  %s",
		data.Repository.FullName,
		data.PullRequest.User.Login,
		data.PullRequest.Title,
		data.PullRequest.HtmlURL,
	))
	return
}
