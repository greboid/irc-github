package main

import (
	"fmt"
	"strings"
)

type githubPushHandler struct{}

func (g *githubPushHandler) tidyPushRefspecs(data *pushhook) {
	data.Refspec = g.tidyRefsHeads(data.Refspec)
	data.Refspec = g.tidyRefsTags(data.Refspec)
	data.Baserefspec = g.tidyRefsHeads(data.Baserefspec)
	data.Baserefspec = g.tidyRefsTags(data.Baserefspec)
}

func (g *githubPushHandler) tidyRefsHeads(input string) string {
	if strings.HasPrefix(input, "refs/heads/") {
		return fmt.Sprintf("branch %s", strings.TrimPrefix(input, "refs/heads/"))
	}
	return input
}

func (g *githubPushHandler) tidyRefsTags(input string) string {
	if strings.HasPrefix(input, "refs/tags/") {
		return fmt.Sprintf("tag %s", strings.TrimPrefix(input, "refs/tags/"))
	}
	return input
}

func (g *githubPushHandler) handlePushEvent(data pushhook) (messages []string) {
	g.tidyPushRefspecs(&data)
	if data.Created {
		return g.handleCreate(data)
	} else if data.Deleted {
		return g.handleDelete(data)
	} else {
		return g.handleCommit(data)
	}
}

func (g *githubPushHandler) handleDelete(data pushhook) (messages []string) {
	messages = append(messages, fmt.Sprintf(
		"[%s] %s deleted %s",
		data.Repository.FullName,
		data.Sender.Login,
		data.Refspec,
	))
	return
}

func (g *githubPushHandler) handleCreate(data pushhook) (messages []string) {
	if data.Baserefspec == "" {
		messages = append(messages, fmt.Sprintf(
			"[%s] %s created %s - %s",
			data.Repository.FullName,
			data.Pusher.Name,
			data.Refspec,
			data.CompareLink,
		))
	} else {
		messages = append(messages, fmt.Sprintf(
			"[%s] %s created %s from %s - %s",
			data.Repository.FullName,
			data.Pusher.Name,
			data.Refspec,
			data.Baserefspec,
			data.CompareLink,
		))
	}
	return
}

func (g *githubPushHandler) handleCommit(data pushhook) (messages []string) {
	messages = append(messages, fmt.Sprintf(
		"[%s] %s pushed %d commits to %s - %s",
		data.Repository.FullName,
		data.Pusher.Name,
		len(data.Commits),
		data.Refspec,
		data.CompareLink,
	))
	for _, commit := range data.Commits {
		messages = append(messages, fmt.Sprintf(
			"[%s] %s committed %s - %s",
			data.Repository.FullName,
			commit.Author.User,
			commit.ID[len(commit.ID)-6:],
			strings.SplitN(commit.Message, "\n", 2)[0],
		))
	}
	return
}
