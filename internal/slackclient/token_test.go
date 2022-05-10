// These tests require that you be logged into Slack on the current machine.
// You must also pass '-tags network' to 'go test'
package slackclient

import "testing"

func TestGetCookie(t *testing.T) {
	cookie, err := getCookie()
	if err != nil {
		t.Error(err)
	}
	t.Error(cookie)

	if cookie == "" {
		t.Error("empty cookie")
	}
}

func TestGetAuth(t *testing.T) {
	auth, err := getSlackAuth("github")
	if err != nil {
		t.Error(err)
	}

	if auth.Token == "" {
		t.Fatal("empty token")
	}

	if auth.Cookies["d"] == "" {
		t.Fatal("empty cookie")
	}
}
