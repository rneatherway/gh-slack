//go:build darwin
// +build darwin

package slackclient

import (
	"os"
	"path"

	"github.com/zalando/go-keyring"
)

func slackConfigDirs() []string {
	home := os.Getenv("HOME")
	return []string{
		path.Join(home, "Library", "Application Support"),
		path.Join(home, "Library", "Containers", "com.tinyspeck.slackmacgap", "Data", "Library", "Application Support"),
	}
}

func cookiePassword() ([]byte, error) {
	secret, err := keyring.Get("Slack Safe Storage", "Slack")
	if err != nil {
		return nil, err
	}

	return []byte(secret), nil
}

func iterations() int {
	return 1003
}
