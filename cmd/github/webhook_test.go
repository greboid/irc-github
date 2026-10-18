package main

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func Test_githubWebhookHandler_handleWebhook(t *testing.T) {
	type args struct {
		eventType string
		filename  string
	}
	tests := []struct {
		name     string
		client   *mockIRCPluginClient
		args     args
		err      bool
		output   int
		finished chan bool
	}{
		{
			name:     "valid ping",
			finished: make(chan bool, 1),
			client:   &mockIRCPluginClient{},
			args: args{
				eventType: "ping",
				filename:  "ping.json",
			},
			err:    false,
			output: 1,
		},
		{
			name:     "valid push",
			finished: make(chan bool, 1),
			client:   &mockIRCPluginClient{},
			args: args{
				eventType: "push",
				filename:  "push/basic.json",
			},
			err:    false,
			output: 1,
		},
		{
			name:   "error push",
			client: &mockIRCPluginClient{},
			args: args{
				eventType: "push",
				filename:  "",
			},
			err:    true,
			output: 0,
		},
		{
			name:     "valid pull_request",
			finished: make(chan bool, 1),
			client:   &mockIRCPluginClient{},
			args: args{
				eventType: "pull_request",
				filename:  "pullrequest_closed_1.json",
			},
			err:    false,
			output: 1,
		},
		{
			name:   "error pull_request",
			client: &mockIRCPluginClient{},
			args: args{
				eventType: "pull_request",
				filename:  "",
			},
			err:    true,
			output: 0,
		},
		{
			name:     "valid issues",
			finished: make(chan bool, 1),
			client:   &mockIRCPluginClient{},
			args: args{
				eventType: "issues",
				filename:  "issues/create_1.json",
			},
			err:    false,
			output: 1,
		},
		{
			name:   "error issues",
			client: &mockIRCPluginClient{},
			args: args{
				eventType: "issues",
				filename:  "",
			},
			err:    true,
			output: 0,
		},
		{
			name:     "valid issue_comment",
			finished: make(chan bool, 1),
			client:   &mockIRCPluginClient{},
			args: args{
				eventType: "issue_comment",
				filename:  "issuecomments/commented_1.json",
			},
			err:    false,
			output: 1,
		},
		{
			name:   "error issue_comment",
			client: &mockIRCPluginClient{},
			args: args{
				eventType: "issue_comment",
				filename:  "",
			},
			err:    true,
			output: 0,
		},
		{
			name:     "valid create",
			finished: make(chan bool, 1),
			client:   &mockIRCPluginClient{},
			args: args{
				eventType: "create",
				filename:  "create_tag_1.json",
			},
			err:    false,
			output: 1,
		},
		{
			name:   "error create",
			client: &mockIRCPluginClient{},
			args: args{
				eventType: "create",
				filename:  "",
			},
			err:    true,
			output: 0,
		},
		{
			name:     "valid release",
			finished: make(chan bool, 1),
			client:   &mockIRCPluginClient{},
			args: args{
				eventType: "release",
				filename:  "releases_published_1.json",
			},
			err:    false,
			output: 1,
		},
		{
			name:   "error release",
			client: &mockIRCPluginClient{},
			args: args{
				eventType: "release",
				filename:  "",
			},
			err:    true,
			output: 0,
		},
		{
			name:     "valid check_run",
			finished: make(chan bool, 1),
			client:   &mockIRCPluginClient{},
			args: args{
				eventType: "check_run",
				filename:  "checkruns/failure_1.json",
			},
			err:    false,
			output: 1,
		},
		{
			name:   "error check_run",
			client: &mockIRCPluginClient{},
			args: args{
				eventType: "check_run",
				filename:  "",
			},
			err:    true,
			output: 0,
		},
		{
			name:     "valid check_suite",
			finished: make(chan bool, 1),
			client:   &mockIRCPluginClient{},
			args: args{
				eventType: "check_suite",
				filename:  "checksuites/failure_1.json",
			},
			err:    false,
			output: 1,
		},
		{
			name:   "error check_suite",
			client: &mockIRCPluginClient{},
			args: args{
				eventType: "check_suite",
				filename:  "",
			},
			err:    true,
			output: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.client.finished = tt.finished
			g := &githubWebhookHandler{
				sender: tt.client,
			}
			var bodyBytes []byte
			if tt.args.filename != "" {
				bodyBytes, _ = getTestDataBytes(tt.args.filename)

			}
			err := g.handleWebhook(tt.args.eventType, bodyBytes)
			if !tt.err {
				select {
				case <-tt.finished:
					break
				case <-time.After(5 * time.Millisecond):
					break
				}
			}
			if (tt.err && err == nil) || (!tt.err && err != nil) {
				t.Errorf("handleWebhook() error = %v, wanted %v", err, tt.err)
			}
			if tt.client.sentChannelMessages != tt.output {
				t.Errorf("handleWebhook() sent = %v, wanted %v", tt.client.sentChannelMessages, tt.output)
			}
		})
	}
}

type mockRecordingIRCPluginClient struct {
	messages []string
}

func (m *mockRecordingIRCPluginClient) SendChannelMessage(_ string, messages ...string) error {
	m.messages = append(m.messages, messages...)
	return nil
}

func Test_rewriteMessage(t *testing.T) {
	rules := []rewriteRule{
		{from: "https://forge.example.com/", to: "https://github.com/"},
		{from: "https://other.example.com/", to: "https://github.com/other/"},
	}
	tests := []struct {
		name    string
		rules   []rewriteRule
		message string
		wanted  string
	}{
		{
			name:    "No rules",
			rules:   nil,
			message: "https://forge.example.com/user/repo",
			wanted:  "https://forge.example.com/user/repo",
		},
		{
			name:    "Single rule match",
			rules:   rules[:1],
			message: "[user/repo] created branch main - https://forge.example.com/user/repo/compare/a...b",
			wanted:  "[user/repo] created branch main - https://github.com/user/repo/compare/a...b",
		},
		{
			name:    "Later rule",
			rules:   rules,
			message: "https://other.example.com/user/repo",
			wanted:  "https://github.com/other/user/repo",
		},
		{
			name:    "Multiple occurrences",
			rules:   rules[:1],
			message: "https://forge.example.com/a https://forge.example.com/b",
			wanted:  "https://github.com/a https://github.com/b",
		},
		{
			name:    "No match",
			rules:   rules,
			message: "https://unrelated.example.com/user/repo",
			wanted:  "https://unrelated.example.com/user/repo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rewriteMessage(tt.rules, tt.message); got != tt.wanted {
				t.Errorf("rewriteMessage() = %q, want %q", got, tt.wanted)
			}
		})
	}
}

func Test_githubWebhookHandler_sendMessage_rewritesLinks(t *testing.T) {
	client := &mockRecordingIRCPluginClient{}
	g := &githubWebhookHandler{
		sender: client,
		rewriteRules: []rewriteRule{
			{from: "https://forge.example.com/", to: "https://github.com/"},
		},
	}
	err := g.sendMessage(false, "[user/repo] user pushed 1 commits to branch main - https://forge.example.com/user/repo/compare/a...b")
	if err != nil {
		t.Fatalf("sendMessage() error = %v", err)
	}
	wanted := "[user/repo] user pushed 1 commits to branch main - https://github.com/user/repo/compare/a...b"
	if len(client.messages) != 1 || client.messages[0] != wanted {
		t.Errorf("sendMessage() sent = %#+v, want %#+v", client.messages, []string{wanted})
	}
}

func Test_githubWebhookHandler_sendMessage(t *testing.T) {
	tests := []struct {
		name     string
		client   *mockFailingIRCPluginClient
		messages []string
		wanted   error
	}{
		{
			name:     "Check working",
			client:   &mockFailingIRCPluginClient{err: false},
			messages: []string{"test"},
			wanted:   nil,
		},
		{
			name:     "Check Failing",
			client:   &mockFailingIRCPluginClient{err: true},
			messages: []string{"test"},
			wanted:   errors.New("fake error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &githubWebhookHandler{
				sender: tt.client,
			}
			got := g.sendMessage(false, tt.messages...)
			if !reflect.DeepEqual(got, tt.wanted) {
				t.Errorf("sendMessage() sent = %v, wanted %v", got, tt.wanted)
			}
		})
	}
}
