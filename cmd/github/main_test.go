package main

import (
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

func getTestData(filename string, output interface{}) error {
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