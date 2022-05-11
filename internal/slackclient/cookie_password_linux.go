//go:build linux
// +build linux

package slackclient

import (
	"errors"

	"r00t2.io/gosecret"
)

func slackConfigDirs() []string {
	if xdgConfigDir, found := os.LookupEnv("XDG_CONFIG_DIR"); found {
		return []string{xdgConfigDir}
	}

	home := os.Getenv("HOME")
	return []string{path.Join(home, ".config")}
}

func cookiePassword() ([]byte, error) {
	service, err := gosecret.NewService()
	if err != nil {
		return nil, err
	}
	defer service.Close()

	itemAttrs := map[string]string{
		"xdg:schema":  "chrome_libsecret_os_crypt_password_v2",
		"application": "Slack",
	}

	unlockedItems, _, err := service.SearchItems(itemAttrs)
	if err != nil {
		return nil, err
	}

	switch len(unlockedItems) {
	case 0:
		return nil, errors.New("no matching unlocked items found")
	case 1:
		return unlockedItems[0].Secret.Value, nil
	default:
		return nil, errors.New("multiple items found")
	}
}

func iterations() int {
	return 1
}
