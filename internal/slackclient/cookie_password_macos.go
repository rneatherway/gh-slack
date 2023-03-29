//go:build darwin
// +build darwin

package slackclient

import (
	"errors"

	"github.com/keybase/go-keychain"
)

func cookiePassword() ([]byte, error) {
	accountNames := []string{"Slack Key", "Slack"}

	var err error
	for _, accountName := range accountNames {
		var password []byte
		password, err = cookiePasswordFromKeychain(accountName)
		if err == nil {
			return password, nil
		}
	}

	return []byte{}, err
}

func cookiePasswordFromKeychain(accountName string) ([]byte, error) {
	query := keychain.NewItem()
	query.SetService("Slack Safe Storage")
	query.SetAccount("Slack Key")
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
