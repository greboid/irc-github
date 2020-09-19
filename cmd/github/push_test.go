package main

import (
	"github.com/greboid/irc/v2/rpc"
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
		args   *pushhook
		want   *pushhook
	}{
		{
			name:   "refspec: master branch",
			fields: fields{},
			args:   &pushhook{Refspec: "refs/heads/master"},
			want:   &pushhook{Refspec: "branch master"},
		},
		{
			name:   "refspec: tag v1",
			fields: fields{},
			args:   &pushhook{Refspec: "refs/tags/v1.0.0"},
			want:   &pushhook{Refspec: "tag v1.0.0"},
		},
		{
			name:   "baserefspec: master branch",
			fields: fields{},
			args:   &pushhook{Baserefspec: "refs/heads/master"},
			want:   &pushhook{Baserefspec: "branch master"},
		},
		{
			name:   "baserefspec: tag v1",
			fields: fields{},
			args:   &pushhook{Baserefspec: "refs/tags/v1.0.0"},
			want:   &pushhook{Baserefspec: "tag v1.0.0"},
		},
		{
			name:   "refspec: non master",
			fields: fields{},
			args:   &pushhook{Baserefspec: "refs/heads/testing"},
			want:   &pushhook{Baserefspec: "branch testing"},
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
	tests := []string{"push/basic.json", "push/tag.json"}
	gold := goldie.New(t)
	for index := range tests {
		t.Run(tests[index], func(t *testing.T) {
			g := &githubPushHandler{}
			hook := pushhook{}
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
			hook := pushhook{}
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
			hook := pushhook{}
			err := getTestData(tests[index], &hook)
			if err != nil {
				t.Fatal("Unable to parse example data")
			}
			got := []byte(strings.Join(g.handleDelete(hook), "\n"))
			gold.Assert(t, tests[index], got)
		})
	}
}
