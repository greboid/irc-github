package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"testing"

	"github.com/sebdah/goldie/v2"
)

func Test_forgejoEventType(t *testing.T) {
	tests := []struct {
		name     string
		headers  http.Header
		wantType string
	}{
		{
			name:     "Forgejo header",
			headers:  http.Header{"X-Forgejo-Event": []string{"push"}},
			wantType: "push",
		},
		{
			name:     "Gitea header fallback",
			headers:  http.Header{"X-Gitea-Event": []string{"issues"}},
			wantType: "issues",
		},
		{
			name:     "Forgejo header preferred over Gitea",
			headers:  http.Header{"X-Forgejo-Event": []string{"push"}, "X-Gitea-Event": []string{"issues"}},
			wantType: "push",
		},
		{
			name:     "No headers",
			headers:  http.Header{},
			wantType: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := forgejoEventType(tt.headers); got != tt.wantType {
				t.Errorf("forgejoEventType() = %q, want %q", got, tt.wantType)
			}
		})
	}
}

func Test_forgejoSignature(t *testing.T) {
	tests := []struct {
		name       string
		headers    http.Header
		wantSecret string
	}{
		{
			name:       "Forgejo header",
			headers:    http.Header{"X-Forgejo-Signature": []string{"abc123"}},
			wantSecret: "abc123",
		},
		{
			name:       "Gitea header fallback",
			headers:    http.Header{"X-Gitea-Signature": []string{"def456"}},
			wantSecret: "def456",
		},
		{
			name:       "Forgejo header preferred over Gitea",
			headers:    http.Header{"X-Forgejo-Signature": []string{"abc123"}, "X-Gitea-Signature": []string{"def456"}},
			wantSecret: "abc123",
		},
		{
			name:       "No headers",
			headers:    http.Header{},
			wantSecret: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := forgejoSignature(tt.headers); got != tt.wantSecret {
				t.Errorf("forgejoSignature() = %q, want %q", got, tt.wantSecret)
			}
		})
	}
}

func Test_CheckForgejoSecret(t *testing.T) {
	body := []byte("test body")
	h := hmac.New(sha256.New, []byte("secret"))
	h.Write(body)
	valid := hex.EncodeToString(h.Sum(nil))
	if !CheckForgejoSecret(body, valid, "secret") {
		t.Error("CheckForgejoSecret() = false for valid signature, want true")
	}
	if CheckForgejoSecret(body, "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", "secret") {
		t.Error("CheckForgejoSecret() = true for invalid signature, want false")
	}
}

func Test_forgejoWebhookHandler_handleWebhook(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		filename  string
		client    *mockIRCPluginClient
		err       bool
		output    int
	}{
		{
			name:      "valid ping",
			eventType: "ping",
			filename:  "forgejo/ping.json",
			client:    &mockIRCPluginClient{},
			err:       false,
			output:    0,
		},
		{
			name:      "valid push",
			eventType: "push",
			filename:  "forgejo/push/basic.json",
			client:    &mockIRCPluginClient{},
			err:       false,
			output:    1,
		},
		{
			name:      "error push",
			eventType: "push",
			filename:  "",
			client:    &mockIRCPluginClient{},
			err:       true,
			output:    0,
		},
		{
			name:      "valid pull_request opened",
			eventType: "pull_request",
			filename:  "forgejo/pullrequest/opened_1.json",
			client:    &mockIRCPluginClient{},
			err:       false,
			output:    1,
		},
		{
			name:      "valid pull_request merged",
			eventType: "pull_request",
			filename:  "forgejo/pullrequest/merged_1.json",
			client:    &mockIRCPluginClient{},
			err:       false,
			output:    1,
		},
		{
			name:      "valid pull_request closed",
			eventType: "pull_request",
			filename:  "forgejo/pullrequest/closed_1.json",
			client:    &mockIRCPluginClient{},
			err:       false,
			output:    1,
		},
		{
			name:      "error pull_request",
			eventType: "pull_request",
			filename:  "",
			client:    &mockIRCPluginClient{},
			err:       true,
			output:    0,
		},
		{
			name:      "valid issues",
			eventType: "issues",
			filename:  "forgejo/issues/create_1.json",
			client:    &mockIRCPluginClient{},
			err:       false,
			output:    1,
		},
		{
			name:      "valid issue_comment",
			eventType: "issue_comment",
			filename:  "forgejo/issuecomments/commented_1.json",
			client:    &mockIRCPluginClient{},
			err:       false,
			output:    1,
		},
		{
			name:      "valid release",
			eventType: "release",
			filename:  "forgejo/releases/published_1.json",
			client:    &mockIRCPluginClient{},
			err:       false,
			output:    1,
		},
		{
			name:      "valid create",
			eventType: "create",
			filename:  "forgejo/create/branch_1.json",
			client:    &mockIRCPluginClient{},
			err:       false,
			output:    1,
		},
		{
			name:      "valid delete",
			eventType: "delete",
			filename:  "forgejo/delete/branch_1.json",
			client:    &mockIRCPluginClient{},
			err:       false,
			output:    1,
		},
		{
			name:      "valid package",
			eventType: "package",
			filename:  "forgejo/packages/container_1.json",
			client:    &mockIRCPluginClient{},
			err:       false,
			output:    1,
		},
		{
			name:      "valid package deleted",
			eventType: "package",
			filename:  "forgejo/packages/deleted_1.json",
			client:    &mockIRCPluginClient{},
			err:       false,
			output:    1,
		},
		{
			name:      "error package",
			eventType: "package",
			filename:  "",
			client:    &mockIRCPluginClient{},
			err:       true,
			output:    0,
		},
		{
			name:      "unknown event",
			eventType: "repository",
			filename:  "forgejo/create/branch_1.json",
			client:    &mockIRCPluginClient{},
			err:       false,
			output:    0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.client.finished = make(chan bool, 1)
			g := &forgejoWebhookHandler{
				gh: &githubWebhookHandler{
					sender: tt.client,
				},
			}
			var bodyBytes []byte
			if tt.filename != "" {
				bodyBytes, _ = getTestDataBytes(tt.filename)
			}
			err := g.handleWebhook(tt.eventType, bodyBytes)
			if (tt.err && err == nil) || (!tt.err && err != nil) {
				t.Errorf("handleWebhook() error = %v, wanted %v", err, tt.err)
			}
			if tt.client.sentChannelMessages != tt.output {
				t.Errorf("handleWebhook() sent = %v, wanted %v", tt.client.sentChannelMessages, tt.output)
			}
		})
	}
}

func Test_forgejoWebhookHandler_sendMessageError(t *testing.T) {
	g := &forgejoWebhookHandler{
		gh: &githubWebhookHandler{
			sender: &mockFailingIRCPluginClient{err: true},
		},
	}
	bodyBytes, _ := getTestDataBytes("forgejo/issues/create_1.json")
	if err := g.handleWebhook("issues", bodyBytes); err != nil {
		t.Errorf("handleWebhook() error = %v, wanted nil", err)
	}
	if g.gh.hookSender != "testuser" {
		t.Errorf("handleWebhook() hookSender = %q, want %q", g.gh.hookSender, "testuser")
	}
}

func Test_forgejo_eventMessages(t *testing.T) {
	tests := []struct {
		eventType string
		filename  string
	}{
		{eventType: "push", filename: "forgejo/push/basic.json"},
		{eventType: "pull_request", filename: "forgejo/pullrequest/opened_1.json"},
		{eventType: "pull_request", filename: "forgejo/pullrequest/merged_1.json"},
		{eventType: "pull_request", filename: "forgejo/pullrequest/closed_1.json"},
		{eventType: "issues", filename: "forgejo/issues/create_1.json"},
		{eventType: "issue_comment", filename: "forgejo/issuecomments/commented_1.json"},
		{eventType: "release", filename: "forgejo/releases/published_1.json"},
		{eventType: "create", filename: "forgejo/create/branch_1.json"},
		{eventType: "delete", filename: "forgejo/delete/branch_1.json"},
		{eventType: "package", filename: "forgejo/packages/container_1.json"},
		{eventType: "package", filename: "forgejo/packages/deleted_1.json"},
	}
	gold := goldie.New(t)
	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			hook := webhook{}
			err := getTestData(tt.filename, &hook)
			if err != nil {
				t.Fatal("Unable to parse example data")
			}
			got := []byte(strings.Join(forgejoEventHandlers[tt.eventType](hook), "\n"))
			gold.Assert(t, tt.filename, got)
		})
	}
}
