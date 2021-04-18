package main

import (
	"encoding/json"
	"fmt"
)

type githubWebhookHandler struct {
}

func (g *githubWebhookHandler) handleWebhook(eventType string, bodyBytes []byte) error {
	switch eventType {
	case "ping":
		data := pinghook{}
		err := json.Unmarshal(bodyBytes, &data)
		if err == nil {
			go func() {
				err := g.sendMessage(data.Repository.IsPrivate, fmt.Sprintf("Ping received for %s", data.Repository.FullName))
				if err != nil {
					log.Errorf("Error handling ping: %s", err.Error())
				}
			}()
		} else {
			log.Errorf("Error handling ping: %s", err.Error())
			return err
		}
	case "push":
		data := pushhook{}
		handler := githubPushHandler{}
		err := json.Unmarshal(bodyBytes, &data)
		if err == nil {
			go func() {
				err := g.sendMessage(data.Repository.IsPrivate, handler.handlePushEvent(data)...)
				if err != nil {
					log.Errorf("Error handling push: %s", err.Error())
				}
			}()
		} else {
			log.Errorf("Error handling push: %s", err.Error())
			return err
		}
	case "pull_request":
		data := prhook{}
		handler := githubPRHandler{}
		err := json.Unmarshal(bodyBytes, &data)
		if err == nil {
			go func() {
				err := g.sendMessage(data.Repository.IsPrivate, handler.handlePREvent(data)...)
				if err != nil {
					log.Errorf("Error handling push: %s", err.Error())
				}
			}()
		} else {
			log.Errorf("Error handling PR: %s", err.Error())
			return err
		}
	case "issues":
		data := issuehook{}
		handler := githubissuehandler{}
		err := json.Unmarshal(bodyBytes, &data)
		if err == nil {
			go func() {
				err := g.sendMessage(data.Repository.IsPrivate, handler.handleIssueEvent(data)...)
				if err != nil {
					log.Errorf("Error handling push: %s", err.Error())
				}
			}()
		} else {
			log.Errorf("Error handling PR: %s", err.Error())
			return err
		}
	case "issue_comment":
		data := issuehook{}
		handler := githubIssueCommenthandler{}
		err := json.Unmarshal(bodyBytes, &data)
		if err == nil {
			go func() {
				err := g.sendMessage(data.Repository.IsPrivate, handler.handleIssueCommentEvent(data)...)
				if err != nil {
					log.Errorf("Error handling push: %s", err.Error())
				}
			}()
		} else {
			log.Errorf("Error handling PR: %s", err.Error())
			return err
		}
	case "check_run":
		// TODO: Handle
		return nil
	case "release":
		// TODO: Handle
		return nil
	case "create":
		// TODO: Handle
		return nil
	case "check_suite":
		// TODO: Handle
		return nil
	}
	return nil
}

func (g *githubWebhookHandler) sendMessage(isPrivate bool, messages ...string) error {
	notifyChannel := *Channel
	if isPrivate && *HidePrivate {
		return nil
	}
	if isPrivate && len(*PrivateChannel) != 0 {
		notifyChannel = *PrivateChannel
	}
	errors := helper.SendChannelMessage(notifyChannel, messages...)
	return errors
}
