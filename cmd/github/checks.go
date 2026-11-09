package main

import "fmt"

type githubCheckRunHandler struct{}

type githubCheckSuiteHandler struct{}

type githubStatusHandler struct{}

// Only completed runs with actionable outcomes are announced; successes would
// duplicate the push notification once per check.
func actionableConclusion(conclusion string) string {
	switch conclusion {
	case "failure":
		return "failed"
	case "error":
		return "errored"
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

func (g *githubStatusHandler) handleStatusEvent(data webhook) (messages []string) {
	conclusion := actionableConclusion(data.State)
	if conclusion == "" {
		return []string{}
	}
	message := fmt.Sprintf("[%s] status check %s %s", data.Repository.FullName, data.Context, conclusion)
	if data.TargetURL != "" {
		message += " - " + data.TargetURL
	}
	return []string{message}
}

func handleForgejoWorkflowRunEvent(data webhook) (messages []string) {
	conclusion := actionableConclusion(data.Run.Status)
	if conclusion == "" {
		return []string{}
	}
	messages = append(messages, fmt.Sprintf(
		"[%s] workflow %s on %s %s - %s",
		data.repo().FullName,
		data.Run.WorkflowID,
		data.Run.PrettyRef,
		conclusion,
		data.Run.HtmlURL,
	))
	return
}

func handleForgejoWorkflowJobEvent(data webhook) (messages []string) {
	conclusion := actionableConclusion(data.Job.Status)
	if conclusion == "" {
		return []string{}
	}
	messages = append(messages, fmt.Sprintf(
		"[%s] workflow %s job %s %s - %s",
		data.repo().FullName,
		data.Run.WorkflowID,
		data.Job.Name,
		conclusion,
		data.Job.HtmlURL,
	))
	return
}
