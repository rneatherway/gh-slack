package slackclient

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestSendMessageToThread(t *testing.T) {
	var sent SendMessage
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if err := json.NewDecoder(req.Body).Decode(&sent); err != nil {
			t.Fatal(err)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"ok":true,"ts":"123.456"}`)),
		}, nil
	})

	client, err := Null("test", transport)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := client.SendMessage("C123", "hello", "100.200"); err != nil {
		t.Fatal(err)
	}

	if sent.Channel != "C123" || sent.Text != "hello" || sent.ThreadTS != "100.200" {
		t.Fatalf("unexpected message: %+v", sent)
	}
}
