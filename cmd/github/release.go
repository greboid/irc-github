package main

import "fmt"

type githubReleaseHandler struct{}

func (g *githubReleaseHandler) handleReleaseEvent(data webhook) (messages []string) {
	switch data.Action {
	case "created":
		return g.handleReleaseCreated(data)
	case "published", "released":
		return g.handleReleasePublished(data)
	case "prereleased":
		return g.handleReleasePrereleased(data)
	}
	return []string{}
}

func (g *githubReleaseHandler) releaseName(data webhook) string {
	if data.Release.Name == "" {
		return data.Release.TagName
	}
	return data.Release.Name
}

func (g *githubReleaseHandler) handleReleaseCreated(data webhook) (messages []string) {
	messages = append(messages, fmt.Sprintf(
		"[%s] %s created release %s - %s",
		data.Repository.FullName,
		data.Sender.Login,
		g.releaseName(data),
		data.Release.HtmlURL,
	))
	return
}

func (g *githubReleaseHandler) handleReleasePublished(data webhook) (messages []string) {
	messages = append(messages, fmt.Sprintf(
		"[%s] %s released %s - %s",
		data.Repository.FullName,
		data.Sender.Login,
		g.releaseName(data),
		data.Release.HtmlURL,
	))
	return
}

func (g *githubReleaseHandler) handleReleasePrereleased(data webhook) (messages []string) {
	messages = append(messages, fmt.Sprintf(
		"[%s] %s released %s (pre-release) - %s",
		data.Repository.FullName,
		data.Sender.Login,
		g.releaseName(data),
		data.Release.HtmlURL,
	))
	return
}
