package main

import "fmt"

type githubCheckRunHandler struct{}

type githubCheckSuiteHandler struct{}

// Only completed runs with actionable outcomes are announced; successes would
// duplicate the push notification once per check.
func actionableConclusion(conclusion string) string {
	switch conclusion {
	case "failure":
		return "failed"
	case "timed_out":
		return "timed out"
	case "cancelled":
		return "was cancelled"
	}
	return ""
}

func (g *githubCheckRunHandler) handleCheckRunEvent(data webhook) (messages []string) {
	if data.Action != "completed" {
		return []string{}
	}
	conclusion := actionableConclusion(data.CheckRun.Conclusion)
	if conclusion == "" {
		return []string{}
	}
	messages = append(messages, fmt.Sprintf(
		"[%s] check run %s %s - %s",
		data.Repository.FullName,
		data.CheckRun.Name,
		conclusion,
		data.CheckRun.HtmlURL,
	))
	return
}

func (g *githubCheckSuiteHandler) handleCheckSuiteEvent(data webhook) (messages []string) {
	if data.Action != "completed" {
		return []string{}
	}
	conclusion := actionableConclusion(data.CheckSuite.Conclusion)
	if conclusion == "" {
		return []string{}
	}
	messages = append(messages, fmt.Sprintf(
		"[%s] %s check suite on branch %s %s",
		data.Repository.FullName,
		data.CheckSuite.App.Name,
		data.CheckSuite.HeadBranch,
		conclusion,
	))
	return
}
