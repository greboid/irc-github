package main

import (
	"github.com/sebdah/goldie/v2"
	"strings"
	"testing"
)

func Test_githubissuehandler_handIssueCommentEvent_comment(t *testing.T) {
	g := &githubIssueCommenthandler{}
	hook := issuehook{}
	err := getTestData("issuecomments/commented_1.json", &hook)
	if err != nil {
		t.Fatal("Unable to parse example data")
	}
	messages := g.handleIssueCommentEvent(hook)
	if len(messages) == 0 {
		t.Fatal("Output expected, none provided")
	}
}

func Test_githubissuehandler_handIssueCommentEvent_unknown(t *testing.T) {
	g := &githubIssueCommenthandler{}
	hook := issuehook{}
	hook.Action = "ThisWillError"
	messages := g.handleIssueCommentEvent(hook)
	if len(messages) != 0 {
		t.Fatal("Output provided, none expected")
	}
}

func Test_githubissuehandler_handleIssueCommentCreated(t *testing.T) {
	tests := []string{"issuecomments/commented_1.json"}
	gold := goldie.New(t)
	for index := range tests {
		t.Run(tests[index], func(t *testing.T) {
			g := &githubIssueCommenthandler{}
			hook := issuehook{}
			err := getTestData(tests[index], &hook)
			if err != nil {
				t.Fatal("Unable to parse example data")
			}
			got := []byte(strings.Join(g.handleIssueCommentCreated(hook), "\n"))
			gold.Assert(t, tests[index], got)
		})

	}
}
