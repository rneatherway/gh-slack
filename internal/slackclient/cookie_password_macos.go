//go:build darwin
// +build darwin

package slackclient

import "github.com/zalando/go-keyring"

func cookiePassword() ([]byte, error) {
	secret, err := keyring.Get("Slack Safe Storage", "Slack")
	if err != nil {
		return nil, err
	}
	return secret, nil
}

func iterations() int {
	return 1003
}
