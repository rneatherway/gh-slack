package extclient

import (
	"fmt"
	"os"
	"path"
	"runtime"
)

func slackConfigDir() string {
	switch runtime.GOOS {
	case "windows":
		return path.Join(os.Getenv("APPDATA"), "Slack")
	case "darwin":
		home := os.Getenv("HOME")
		first := path.Join(home, "Library", "Application Support", "Slack")
		second := path.Join(home, "Library", "Containers", "com.tinyspeck.slackmacgap", "Data", "Library", "Application Support", "Slack")

		if _, err := os.Stat(first); err == nil {
			return first
		}
		return second
	case "linux":
		if xdgConfigDir, found := os.LookupEnv("XDG_CONFIG_DIR"); found {
			return path.Join(xdgConfigDir, "Slack")
		}
		return path.Join(os.Getenv("HOME"), ".config", "Slack")
	default:
		panic(fmt.Sprintf("Platform %q not supported", runtime.GOOS))
	}
}
