package slackclient

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"

	_ "modernc.org/sqlite"
)

type SlackAuth struct {
	Token   string
	Cookies map[string]string
}

var stmt = "SELECT value FROM cookies WHERE host_key=\".slack.com\" AND name=\"d\""

func getCookie() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	var config string
	switch runtime.GOOS {
	case "darwin":
		config = path.Join(home, "Library", "Application Support")
	case "linux":
		config = path.Join(home, ".config")
	default:
		return "", fmt.Errorf("unsupported platform %q", runtime.GOOS)
	}
	cookies := path.Join(config, "Slack", "Cookies")

	db, err := sql.Open("sqlite", cookies)
	if err != nil {
		return "", err
	}

	var cookie string
	err = db.QueryRow(stmt).Scan(&cookie)
	if err != nil {
		return "", err
	}

	return cookie, nil
}

var apiTokenRE = regexp.MustCompile("\"api_token\":\"([^\"]+)\"")

func getSlackAuth(team string) (*SlackAuth, error) {
	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequest("GET", fmt.Sprintf("https://%s.slack.com", team), nil)
	if err != nil {
		return nil, err
	}

	r.AddCookie(&http.Cookie{Name: "d", Value: cookie})

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	matches := apiTokenRE.FindSubmatch(bs)
	if matches == nil {
		return nil, errors.New("api token not found")
	}

	return &SlackAuth{Token: string(matches[1]), Cookies: map[string]string{"d": cookie}}, nil
}

func getSlackAuthFromEnv() (*SlackAuth, error) {
	slackAuth := os.Getenv("SLACK_AUTH")
	if slackAuth == "" {
		return nil, errors.New("environment variable SLACK_AUTH not set")
	}

	token, cookie, found := strings.Cut(slackAuth, "\n")
	if !found {
		return nil, errors.New("environment variable SLACK_AUTH not in expected format")
	}

	key, value, found := strings.Cut(cookie, "=")
	if !found {
		return nil, errors.New("environment variable SLACK_AUTH not in expected format")
	}

	cookie, err := url.PathUnescape(value)
	if err != nil {
		return nil, fmt.Errorf("failed to unescape cookie value: %w", err)
	}

	return &SlackAuth{
		Token:   token,
		Cookies: map[string]string{key: cookie},
	}, nil
}
