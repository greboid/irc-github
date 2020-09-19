package main

import (
	"github.com/sebdah/goldie/v2"
	"strings"
	"testing"
)

func Test_githubissuehandler_handIssueEvent_open(t *testing.T) {
	g := &githubissuehandler{}
	hook := issuehook{}
	err := getTestData("issues/create_1.json", &hook)
	if err != nil {
		t.Fatal("Unable to parse example data")
	}
	messages := g.handleIssueEvent(hook)
	if len(messages) == 0 {
		t.Fatal("Output expected, none provided")
	}
}

func Test_githubissuehandler_handIssueEvent_closed(t *testing.T) {
	g := &githubissuehandler{}
	hook := issuehook{}
	err := getTestData("issues/closed_1.json", &hook)
	if err != nil {
		t.Fatal("Unable to parse example data")
	}
	messages := g.handleIssueEvent(hook)
	if len(messages) == 0 {
		t.Fatal("Output expected, none provided")
	}
}

func Test_githubissuehandler_handIssueEvent_unknown(t *testing.T) {
	g := &githubissuehandler{}
	hook := issuehook{}
	hook.Action = "ThisWillError"
	messages := g.handleIssueEvent(hook)
	if len(messages) != 0 {
		t.Fatal("Output provided, none expected")
	}
}

func Test_githubissuehandler_handleIssueOpened(t *testing.T) {
	tests := []string{"issues/create_1.json"}
	gold := goldie.New(t)
	for index := range tests {
		t.Run(tests[index], func(t *testing.T) {
			g := &githubissuehandler{}
			hook := issuehook{}
			err := getTestData(tests[index], &hook)
			if err != nil {
				t.Fatal("Unable to parse example data")
			}
			got := []byte(strings.Join(g.handleIssueOpened(hook), "\n"))
			gold.Assert(t, tests[index], got)
		})

	}
}

func Test_githubissuehandler_handleIssueClosed(t *testing.T) {
	tests := []string{"issues/closed_1.json"}
	gold := goldie.New(t)
	for index := range tests {
		t.Run(tests[index], func(t *testing.T) {
			g := &githubissuehandler{}
			hook := issuehook{}
			err := getTestData(tests[index], &hook)
			if err != nil {
				t.Fatal("Unable to parse example data")
			}
			got := []byte(strings.Join(g.handleIssueClosed(hook), "\n"))
			gold.Assert(t, tests[index], got)
		})

	}
}
