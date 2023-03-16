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

	"github.com/rneatherway/gh-slack/internal/httpclient"
	_ "modernc.org/sqlite"
)

type SlackAuth struct {
	Token   string
	Cookies map[string]string
}

var stmt = "SELECT value, encrypted_value FROM cookies WHERE host_key=\".slack.com\" AND name=\"d\""

type CookieDecryptor interface {
	Decrypt(value, key []byte) ([]byte, error)
}

func decrypt(encryptedValue, key []byte) ([]byte, error) {
	switch runtime.GOOS {
	case "windows":
		return WindowsDecryptor{}.Decrypt(encryptedValue, key)
	case "darwin":
		return UnixCookieDecryptor{rounds: 1003}.Decrypt(encryptedValue, key)
	case "linux":
		return UnixCookieDecryptor{rounds: 1}.Decrypt(encryptedValue, key)
	default:
		panic(fmt.Sprintf("platform %q not supported", runtime.GOOS))
	}
}

func getCookie() (string, error) {
	cookieDBFile := path.Join(slackConfigDir(), "Cookies")
	if runtime.GOOS == "windows" {
		cookieDBFile = path.Join(slackConfigDir(), "Network", "Cookies")
	}

	stat, err := os.Stat(cookieDBFile)
	if err != nil {
		return "", fmt.Errorf("could not access Slack cookie database: %w", err)
	}
	if stat.IsDir() {
		return "", fmt.Errorf("directory found at expected Slack cookie database location %q", cookieDBFile)
	}

	if cookieDBFile == "" {
		return "", errors.New("no Slack cookie database found. Are you definitely logged in?")
	}

	db, err := sql.Open("sqlite", cookieDBFile)
	if err != nil {
		return "", err
	}

	var cookie string
	var encryptedValue []byte
	err = db.QueryRow(stmt).Scan(&cookie, &encryptedValue)
	if err != nil {
		return "", err
	}

	if cookie != "" {
		return cookie, nil
	}

	// Remove the version number e.g. v11
	encryptedValue = encryptedValue[3:]

	// We need to decrypt the cookie.
	key, err := cookiePassword()
	if err != nil {
		return "", fmt.Errorf("failed to get cookie password: %w", err)
	}

	decryptedValue, err := decrypt(encryptedValue, key)
	if err != nil {
		return "", err
	}

	return string(decryptedValue), err
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

	resp, err := httpclient.Client.Do(r)
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

var errorNoSlackAuth = errors.New("environment variable SLACK_AUTH not in expected format. Try cloning https://github.com/chrisgavin/slacktoken and running `export SLACK_AUTH=\"$(python3 -m slacktoken get --workspace github)\"`")

func getSlackAuthFromEnv() (*SlackAuth, error) {
	slackAuth := os.Getenv("SLACK_AUTH")
	if slackAuth == "" {
		return nil, errorNoSlackAuth
	}

	token, cookie, found := strings.Cut(slackAuth, "\n")
	if !found {
		return nil, errorNoSlackAuth
	}

	key, value, found := strings.Cut(cookie, "=")
	if !found {
		return nil, errorNoSlackAuth
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
