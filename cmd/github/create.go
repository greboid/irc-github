package main

import "fmt"

type githubCreateHandler struct{}

func (g *githubCreateHandler) handleCreateEvent(data webhook) (messages []string) {
	switch data.RefType {
	case "branch", "tag":
		return g.handleRefCreated(data)
	}
	return []string{}
}

func (g *githubCreateHandler) handleRefCreated(data webhook) (messages []string) {
	messages = append(messages, fmt.Sprintf(
		"[%s] %s created %s %s",
		data.Repository.FullName,
		data.Sender.Login,
		data.RefType,
		data.Refspec,
	))
	return
}
