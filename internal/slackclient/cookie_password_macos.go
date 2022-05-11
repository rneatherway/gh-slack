//go:build darwin
// +build darwin

package slackclient

import (
	"errors"
	"os"
	"path"

	"github.com/keybase/go-keychain"
)

func slackConfigDirs() []string {
	home := os.Getenv("HOME")
	return []string{
		path.Join(home, "Library", "Application Support"),
		path.Join(home, "Library", "Containers", "com.tinyspeck.slackmacgap", "Data", "Library", "Application Support"),
	}
}

func cookiePassword() ([]byte, error) {
	query := keychain.NewItem()
	query.SetSecClass(keychain.SecClassGenericPassword)
	query.SetService("Slack Safe Storage")
	query.SetAccount("Slack")
	query.SetMatchLimit(keychain.MatchLimitOne)
	query.SetReturnAttributes(true)
	query.SetReturnData(true)
	results, err := keychain.QueryItem(query)
	if err != nil {
		return nil, err
	}

	switch len(results) {
	case 0:
		return nil, errors.New("no matching unlocked items found")
	case 1:
		return results[0].Data, nil
	default:
		return nil, errors.New("multiple items found")
	}
}

func iterations() int {
	return 1003
}
