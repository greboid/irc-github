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
