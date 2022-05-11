package slackclient

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
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

	"golang.org/x/crypto/pbkdf2"

	_ "modernc.org/sqlite"
)

type SlackAuth struct {
	Token   string
	Cookies map[string]string
}

var stmt = "SELECT value, encrypted_value FROM cookies WHERE host_key=\".slack.com\" AND name=\"d\""

type CookieDecryptor interface {
	Password() string
}

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
	var encrypted_value []byte
	err = db.QueryRow(stmt).Scan(&cookie, &encrypted_value)
	if err != nil {
		return "", err
	}

	if cookie != "" {
		return cookie, nil
	}

	// We need to decrypt the cookie.

	key, err := cookiePassword()
	if err != nil {
		return "", fmt.Errorf("failed to get cookie password: %w", err)
	}
	dk := pbkdf2.Key(key, []byte("saltysalt"), iterations(), 16, sha1.New)

	block, err := aes.NewCipher(dk)
	if err != nil {
		return "", err
	}

	iv := make([]byte, 16)
	for i := range iv {
		iv[i] = ' '
	}

	mode := cipher.NewCBCDecrypter(block, iv)

	encrypted_value = encrypted_value[3:]
	mode.CryptBlocks(encrypted_value, encrypted_value)

	bytesToStrip := int(encrypted_value[len(encrypted_value)-1])

	return string(encrypted_value[:len(encrypted_value)-bytesToStrip]), nil
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
