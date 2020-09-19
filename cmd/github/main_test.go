package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/greboid/irc/v2/rpc"
	"google.golang.org/grpc"
	"io/ioutil"
)

type mockFailingIRCPluginClient struct {
	err bool
}

func (m *mockFailingIRCPluginClient) SendChannelMessage(context.Context, *rpc.ChannelMessage, ...grpc.CallOption) (*rpc.Error, error) {
	if m.err {
		return &rpc.Error{Message: "Fake error"}, errors.New("fake error")
	} else {
		return nil, nil
	}
}

func (m *mockFailingIRCPluginClient) SendRawMessage(context.Context, *rpc.RawMessage, ...grpc.CallOption) (*rpc.Error, error) {
	if m.err {
		return &rpc.Error{Message: "Fake error"}, errors.New("fake error")
	} else {
		return nil, nil
	}
}

func (m *mockFailingIRCPluginClient) GetMessages(context.Context, *rpc.Channel, ...grpc.CallOption) (rpc.IRCPlugin_GetMessagesClient, error) {
	if m.err {
		return nil, errors.New("fake error")
	} else {
		return nil, nil
	}
}

func (m *mockFailingIRCPluginClient) Ping(context.Context, *rpc.Empty, ...grpc.CallOption) (*rpc.Empty, error) {
	if m.err {
		return &rpc.Empty{}, errors.New("fake error")
	} else {
		return nil, nil
	}
}

type mockIRCPluginClient struct {
	sentRawMessages     int
	sentChannelMessages int
	finished            chan bool
}

func (m *mockIRCPluginClient) SendChannelMessage(context.Context, *rpc.ChannelMessage, ...grpc.CallOption) (*rpc.Error, error) {
	m.sentChannelMessages = m.sentChannelMessages + 1
	m.finished <- true
	return nil, nil
}

func (m *mockIRCPluginClient) SendRawMessage(context.Context, *rpc.RawMessage, ...grpc.CallOption) (*rpc.Error, error) {
	m.sentRawMessages = m.sentRawMessages + 1
	m.finished <- true
	return nil, nil
}

func (m *mockIRCPluginClient) GetMessages(context.Context, *rpc.Channel, ...grpc.CallOption) (rpc.IRCPlugin_GetMessagesClient, error) {
	return nil, nil
}

func (m *mockIRCPluginClient) Ping(context.Context, *rpc.Empty, ...grpc.CallOption) (*rpc.Empty, error) {
	return nil, nil
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
