package main

import (
	"github.com/sebdah/goldie/v2"
	"reflect"
	"strings"
	"testing"
)

func Test_github_handleReleaseEvent(t *testing.T) {
	tests := []string{"releases_created_1.json", "releases_published_1.json", "releases_released_1.json"}
	gold := goldie.New(t)
	for index := range tests {
		t.Run(tests[index], func(t *testing.T) {
			g := &githubReleaseHandler{}
			hook := webhook{}
			err := getTestData(tests[index], &hook)
			if err != nil {
				t.Fatal("Unable to parse example data")
			}
			got := []byte(strings.Join(g.handleReleaseEvent(hook), "\n"))
			gold.Assert(t, tests[index], got)
		})
	}
}

func Test_github_handleReleaseEvent_unhandledAction(t *testing.T) {
	g := &githubReleaseHandler{}
	got := g.handleReleaseEvent(webhook{Action: "edited"})
	if len(got) != 0 {
		t.Errorf("handleReleaseEvent() = %v, want no messages", got)
	}
}

func Test_github_handleReleaseEvent_emptyNameFallsBackToTag(t *testing.T) {
	g := &githubReleaseHandler{}
	got := g.handleReleaseEvent(webhook{
		Action:     "published",
		Repository: Repository{FullName: "greboid/irc"},
		Sender:     Sender{Login: "greboid"},
		Release:    Release{TagName: "v1.2.0", HtmlURL: "https://github.com/greboid/irc/releases/tag/v1.2.0"},
	})
	want := []string{"[greboid/irc] greboid released v1.2.0 - https://github.com/greboid/irc/releases/tag/v1.2.0"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("handleReleaseEvent() = %v, want %v", got, want)
	}
}
