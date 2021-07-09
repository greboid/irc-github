package main

import (
	"encoding/json"
	webhook "github.com/go-playground/webhooks/v6/github"
	"github.com/sebdah/goldie/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_handlePush(t *testing.T) {
	tests := []string{
		"push/commit/basic.json",
		"push/commit/external_merge.json",
		"push/commit/multiline_commit_message.json",
		"push/delete/delete_1.json",
		"push/basic.json",
		"push/tag.json",
	}
	gold := goldie.New(t)

	for i := range tests {
		t.Run(tests[i], func(t *testing.T) {
			s := webhook.PushPayload{}
			g := &githubWebhookHandler{}
			b, err := os.ReadFile(filepath.Join("testdata", tests[i]))
			if err != nil {
				t.Fatal("Unable to parse test data")
			}
			err = json.Unmarshal(b, &s)
			if err != nil {
				t.Fatal("Unable to unmarshal test data")
			}
			//TODO: Should probably check isPrivate
			_, actualM := g.handlePush(s)
			gold.Assert(t, tests[i],  []byte(strings.Join(actualM, "\n")))
		})
	}
}