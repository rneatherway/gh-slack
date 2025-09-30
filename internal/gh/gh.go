package gh

import (
	"fmt"
	"os"

	"github.com/cli/go-gh/v2"
)

func NewIssue(repoUrl string, channelName, content string) error {
	out, _, err := gh.Exec(
		"issue",
		"-R",
		repoUrl,
		"create",
		"--title",
		fmt.Sprintf("Slack conversation archive of `#%s`", channelName),
		"--body",
		content)
	os.Stdout.Write(out.Bytes())
	return err
}

func AddComment(subCmd, url, content string) error {
	out, _, err := gh.Exec(
		subCmd,
		"comment",
		url,
		"--body",
		content)
	os.Stdout.Write(out.Bytes())
	return err
}
