package main

import (
	"github.com/sebdah/goldie/v2"
	"strings"
	"testing"
)

func Test_github_handleCreateEvent(t *testing.T) {
	tests := []string{"create_tag_1.json", "push_create_1.json"}
	gold := goldie.New(t)
	for index := range tests {
		t.Run(tests[index], func(t *testing.T) {
			g := &githubCreateHandler{}
			hook := webhook{}
			err := getTestData(tests[index], &hook)
			if err != nil {
				t.Fatal("Unable to parse example data")
			}
			got := []byte(strings.Join(g.handleCreateEvent(hook), "\n"))
			gold.Assert(t, tests[index], got)
		})
	}
}

func Test_github_handleCreateEvent_unhandledRefType(t *testing.T) {
	g := &githubCreateHandler{}
	got := g.handleCreateEvent(webhook{RefType: "repository"})
	if len(got) != 0 {
		t.Errorf("handleCreateEvent() = %v, want no messages", got)
	}
}
