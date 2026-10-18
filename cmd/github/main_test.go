package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"
)

type mockFailingIRCPluginClient struct {
	err bool
}

func (m *mockFailingIRCPluginClient) SendChannelMessage(_ string, _ ...string) error {
	if m.err {
		return errors.New("fake error")
	} else {
		return nil
	}
}

type mockIRCPluginClient struct {
	sentRawMessages     int
	sentChannelMessages int
	finished            chan bool
}

func (m *mockIRCPluginClient) SendChannelMessage(_ string, _ ...string) error {
	m.sentChannelMessages++
	m.finished <- true
	return nil
}

func getTestData(filename string, output any) error {
	data, err := getTestDataBytes(filename)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &output)
	return err
}

func getTestDataBytes(filename string) ([]byte, error) {
	return ioutil.ReadFile(fmt.Sprintf("./testdata/%s", filename))
}

func Test_parseIgnoredUsers(t *testing.T) {
	tests := []struct {
		name             string
		inputUsers       string
		wantIgnoredUsers []string
	}{
		{
			name:             "Empty string",
			inputUsers:       "",
			wantIgnoredUsers: []string{},
		},
		{
			name:             "Single user",
			inputUsers:       "user",
			wantIgnoredUsers: []string{"user"},
		},
		{
			name:             "Multiple users",
			inputUsers:       "user1,user2",
			wantIgnoredUsers: []string{"user1", "user2"},
		},
		{
			name:             "Nonsense user",
			inputUsers:       "				",
			wantIgnoredUsers: []string{},
		},
		{
			name:             "User trailing comma",
			inputUsers:       "user1, ",
			wantIgnoredUsers: []string{"user1"},
		},
		{
			name:             "User with leading comma",
			inputUsers:       ",user1",
			wantIgnoredUsers: []string{"user1"},
		},
		{
			name:             "Users with spaces",
			inputUsers:       "user1, user2",
			wantIgnoredUsers: []string{"user1", "user2"},
		},
		{
			name:             "Users with blank",
			inputUsers:       "user1,,user2",
			wantIgnoredUsers: []string{"user1", "user2"},
		},
		{
			name:             "Users with rubbish middle",
			inputUsers:       "user1,		,,user2",
			wantIgnoredUsers: []string{"user1", "user2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotIgnoredUsers := parseIgnoredUsers(tt.inputUsers); !reflect.DeepEqual(gotIgnoredUsers, tt.wantIgnoredUsers) {
				t.Errorf("parseIgnoredUsers() = %#+v, want %#+v", gotIgnoredUsers, tt.wantIgnoredUsers)
			}
		})
	}
}

func Test_parseRewriteRules(t *testing.T) {
	tests := []struct {
		name       string
		inputRules string
		wantRules  []rewriteRule
	}{
		{
			name:       "Empty string",
			inputRules: "",
			wantRules:  []rewriteRule{},
		},
		{
			name:       "Single rule",
			inputRules: "https://forge.example.com/=https://github.com/",
			wantRules:  []rewriteRule{{from: "https://forge.example.com/", to: "https://github.com/"}},
		},
		{
			name:       "Multiple rules",
			inputRules: "https://forge.example.com/=https://github.com/,https://other.example.com/=https://github.com/other",
			wantRules: []rewriteRule{
				{from: "https://forge.example.com/", to: "https://github.com/"},
				{from: "https://other.example.com/", to: "https://github.com/other"},
			},
		},
		{
			name:       "Rules with spaces",
			inputRules: "https://a.example/=https://b.example/, https://c.example/=https://d.example/",
			wantRules: []rewriteRule{
				{from: "https://a.example/", to: "https://b.example/"},
				{from: "https://c.example/", to: "https://d.example/"},
			},
		},
		{
			name:       "Rules with blank middle",
			inputRules: "https://a.example/=https://b.example/,,https://c.example/=https://d.example/",
			wantRules: []rewriteRule{
				{from: "https://a.example/", to: "https://b.example/"},
				{from: "https://c.example/", to: "https://d.example/"},
			},
		},
		{
			name:       "Rule without separator",
			inputRules: "https://a.example/",
			wantRules:  []rewriteRule{},
		},
		{
			name:       "Rule with empty source",
			inputRules: "=https://b.example/",
			wantRules:  []rewriteRule{},
		},
		{
			name:       "Rule with empty target",
			inputRules: "https://a.example/=",
			wantRules:  []rewriteRule{},
		},
		{
			name:       "Rule with equals in target",
			inputRules: "https://a.example/=https://b.example/?x=1",
			wantRules:  []rewriteRule{{from: "https://a.example/", to: "https://b.example/?x=1"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotRules := parseRewriteRules(tt.inputRules); !reflect.DeepEqual(gotRules, tt.wantRules) {
				t.Errorf("parseRewriteRules() = %#+v, want %#+v", gotRules, tt.wantRules)
			}
		})
	}
}

func Test_parseSignatureHeader(t *testing.T) {
	tests := []struct {
		name        string
		headerValue string
		wantValue   string
		wantOK      bool
	}{
		{
			name:        "Valid signature",
			headerValue: "sha1=abcdef",
			wantValue:   "abcdef",
			wantOK:      true,
		},
		{
			name:        "Missing separator",
			headerValue: "sha1",
			wantValue:   "",
			wantOK:      false,
		},
		{
			name:        "Wrong algorithm",
			headerValue: "sha256=abcdef",
			wantValue:   "",
			wantOK:      false,
		},
		{
			name:        "Empty header",
			headerValue: "",
			wantValue:   "",
			wantOK:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, gotOK := parseSignatureHeader(tt.headerValue)
			if gotValue != tt.wantValue || gotOK != tt.wantOK {
				t.Errorf("parseSignatureHeader() = %q, %v, want %q, %v", gotValue, gotOK, tt.wantValue, tt.wantOK)
			}
		})
	}
}

func Test_CheckGithubSecret(t *testing.T) {
	body := []byte("test body")
	h := hmac.New(sha1.New, []byte("secret"))
	h.Write(body)
	valid := hex.EncodeToString(h.Sum(nil))
	if !CheckGithubSecret(body, valid, "secret") {
		t.Error("CheckGithubSecret() = false for valid signature, want true")
	}
	if CheckGithubSecret(body, "0123456789012345678901234567890123456789", "secret") {
		t.Error("CheckGithubSecret() = true for invalid signature, want false")
	}
}
