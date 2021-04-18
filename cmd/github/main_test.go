package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
)

type mockFailingIRCPluginClient struct {
	err bool
}

func (m *mockFailingIRCPluginClient) SendChannelMessage(_ string, _ ... string) error {
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

func (m *mockIRCPluginClient) SendChannelMessage(_ string, _ ... string) error {
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
