package mocks

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/rneatherway/gh-slack/internal/slackclient"
)

// MockClient is the mock client
type MockClient struct {
	Next func(*http.Request) (*http.Response, error)
}

func (m *MockClient) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.Next(req)
}

func (m *MockClient) MockSuccessfulAuthResponse() {
	m.Next = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"api_token":"8675309"}`))),
		}, nil
	}
}

func (m *MockClient) MockSuccessfulUsersResponse(fakeUsers []slackclient.User) {
	json := `{"Ok":true,"Members":[`
	for i, user := range fakeUsers {
		json += fmt.Sprintf(`{"ID":"%s","Name":"%s"}`, user.ID, user.Name)
		if i < len(fakeUsers)-1 {
			json += ","
		}
	}
	json += `]}`
	m.Next = func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(json)))}, nil
	}
}
