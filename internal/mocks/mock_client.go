package mocks

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/rneatherway/gh-slack/internal/slackclient"
)

// MockClient is the mock client
type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

var (
	// GetDoFunc fetches the mock client's `Do` func
	GetDoFunc func(req *http.Request) (*http.Response, error)
)

// Do is the mock client's `Do` func
func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return GetDoFunc(req)
}

func MockSuccessfulAuthResponse() {
	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"api_token":"8675309"}`))),
		}, nil
	}
}

func MockSuccessfulUsersResponse(fakeUsers []slackclient.User) {
	json := `{"Ok":true,"Members":[`
	for i, user := range fakeUsers {
		json += fmt.Sprintf(`{"ID":"%s","Name":"%s"}`, user.ID, user.Name)
		if i < len(fakeUsers)-1 {
			json += ","
		}
	}
	json += `]}`
	fmt.Println("JSON:", json)
	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(json)))}, nil
	}
}
