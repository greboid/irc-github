package main

import (
	"github.com/greboid/irc-bot/v5/rpc"
	"github.com/sebdah/goldie/v2"
	"reflect"
	"strings"
	"testing"
)

func Test_github_tidyPushRefspecs(t *testing.T) {
	type fields struct {
		client rpc.IRCPluginClient
	}
	tests := []struct {
		name   string
		fields fields
		args   *webhook
		want   *webhook
	}{
		{
			name:   "refspec: master branch",
			fields: fields{},
			args:   &webhook{Refspec: "refs/heads/master"},
			want:   &webhook{Refspec: "branch master"},
		},
		{
			name:   "refspec: tag v1",
			fields: fields{},
			args:   &webhook{Refspec: "refs/tags/v1.0.0"},
			want:   &webhook{Refspec: "tag v1.0.0"},
		},
		{
			name:   "baserefspec: master branch",
			fields: fields{},
			args:   &webhook{Baserefspec: "refs/heads/master"},
			want:   &webhook{Baserefspec: "branch master"},
		},
		{
			name:   "baserefspec: tag v1",
			fields: fields{},
			args:   &webhook{Baserefspec: "refs/tags/v1.0.0"},
			want:   &webhook{Baserefspec: "tag v1.0.0"},
		},
		{
			name:   "refspec: non master",
			fields: fields{},
			args:   &webhook{Baserefspec: "refs/heads/testing"},
			want:   &webhook{Baserefspec: "branch testing"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &githubPushHandler{}
			if g.tidyPushRefspecs(tt.args); !reflect.DeepEqual(tt.args.Refspec, tt.want.Refspec) {
				t.Errorf("%v != %v", tt.args.Refspec, tt.want.Refspec)
			}
		})
	}
}

func Test_github_handlePushEvent(t *testing.T) {
	tests := []string{"push/basic.json", "push/tag.json", "push/branch.json"}
	gold := goldie.New(t)
	for index := range tests {
		t.Run(tests[index], func(t *testing.T) {
			g := &githubPushHandler{}
			hook := webhook{}
			err := getTestData(tests[index], &hook)
			if err != nil {
				t.Fatal("Unable to parse example data")
			}
			got := []byte(strings.Join(g.handlePushEvent(hook), "\n"))
			gold.Assert(t, tests[index], got)
		})
	}
}

func Test_github_handleCommit(t *testing.T) {
	tests := []string{"push/commit/basic.json", "push/commit/multiline_commit_message.json", "push/commit/external_merge.json"}
	gold := goldie.New(t)
	for index := range tests {
		t.Run(tests[index], func(t *testing.T) {
			g := &githubPushHandler{}
			hook := webhook{}
			err := getTestData(tests[index], &hook)
			if err != nil {
				t.Fatal("Unable to parse example data")
			}
			got := []byte(strings.Join(g.handleCommit(hook), "\n"))
			gold.Assert(t, tests[index], got)
		})
	}
}

func Test_githubPushHandler_handleDelete(t *testing.T) {
	tests := []string{"push/delete/delete_1.json"}
	gold := goldie.New(t)
	for index := range tests {
		t.Run(tests[index], func(t *testing.T) {
			g := &githubPushHandler{}
			hook := webhook{}
			err := getTestData(tests[index], &hook)
			if err != nil {
				t.Fatal("Unable to parse example data")
			}
			got := []byte(strings.Join(g.handleDelete(hook), "\n"))
			gold.Assert(t, tests[index], got)
		})
	}
}

func Test_github_handleCommit_authorDisplayName(t *testing.T) {
	tests := []struct {
		name   string
		commit Commit
		want   string
	}{
		{
			name:   "github username",
			commit: Commit{Author: Author{User: "testuser", Name: "Test User"}},
			want:   "testuser",
		},
		{
			name:   "forgejo automerge with no matching user",
			commit: Commit{Author: Author{User: "", Name: "Auto Updater"}},
			want:   "Auto Updater",
		},
		{
			name:   "both empty",
			commit: Commit{Author: Author{User: "", Name: ""}},
			want:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.commit.authorDisplayName(); got != tt.want {
				t.Errorf("authorDisplayName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func Test_github_handleCommit_shortID(t *testing.T) {
	g := &githubPushHandler{}
	got := g.handleCommit(webhook{
		Repository: Repository{FullName: "user/repo"},
		Commits: []Commit{
			{ID: "abc", Message: "short id\nsecond line"},
		},
	})
	want := []string{
		"[user/repo]  pushed 1 commits to  - ",
		"[user/repo]  committed abc - short id",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("handleCommit() = %v, want %v", got, want)
	}
}
