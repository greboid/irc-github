package main

import (
	"github.com/sebdah/goldie/v2"
	"strings"
	"testing"
)

func Test_github_handleCheckRunEvent(t *testing.T) {
	tests := []string{"checkruns/failure_1.json"}
	gold := goldie.New(t)
	for index := range tests {
		t.Run(tests[index], func(t *testing.T) {
			g := &githubCheckRunHandler{}
			hook := webhook{}
			err := getTestData(tests[index], &hook)
			if err != nil {
				t.Fatal("Unable to parse example data")
			}
			got := []byte(strings.Join(g.handleCheckRunEvent(hook), "\n"))
			gold.Assert(t, tests[index], got)
		})
	}
}

func Test_github_handleCheckRunEvent_notAnnounced(t *testing.T) {
	tests := []struct {
		name string
		data webhook
	}{
		{
			name: "Successful run",
			data: webhook{Action: "completed", CheckRun: CheckRun{Name: "build", Conclusion: "success"}},
		},
		{
			name: "Skipped run",
			data: webhook{Action: "completed", CheckRun: CheckRun{Name: "build", Conclusion: "skipped"}},
		},
		{
			name: "Run in progress",
			data: webhook{Action: "in_progress", CheckRun: CheckRun{Name: "build", Conclusion: ""}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &githubCheckRunHandler{}
			if got := g.handleCheckRunEvent(tt.data); len(got) != 0 {
				t.Errorf("handleCheckRunEvent() = %v, want no messages", got)
			}
		})
	}
}

func Test_github_handleCheckSuiteEvent(t *testing.T) {
	tests := []string{"checksuites/failure_1.json"}
	gold := goldie.New(t)
	for index := range tests {
		t.Run(tests[index], func(t *testing.T) {
			g := &githubCheckSuiteHandler{}
			hook := webhook{}
			err := getTestData(tests[index], &hook)
			if err != nil {
				t.Fatal("Unable to parse example data")
			}
			got := []byte(strings.Join(g.handleCheckSuiteEvent(hook), "\n"))
			gold.Assert(t, tests[index], got)
		})
	}
}

func Test_github_handleCheckSuiteEvent_notAnnounced(t *testing.T) {
	tests := []struct {
		name string
		data webhook
	}{
		{
			name: "Successful suite",
			data: webhook{Action: "completed", CheckSuite: CheckSuite{HeadBranch: "master", Conclusion: "success"}},
		},
		{
			name: "Suite in progress",
			data: webhook{Action: "requested", CheckSuite: CheckSuite{HeadBranch: "master", Conclusion: ""}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &githubCheckSuiteHandler{}
			if got := g.handleCheckSuiteEvent(tt.data); len(got) != 0 {
				t.Errorf("handleCheckSuiteEvent() = %v, want no messages", got)
			}
		})
	}
}
