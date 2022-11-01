package gh

import (
	"fmt"
	"os"

	"github.com/cli/go-gh"
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

func AddComment(issueUrl string, content string) error {
	out, _, err := gh.Exec(
		"issue",
		"comment",
		issueUrl,
		"--body",
		content)
	os.Stdout.Write(out.Bytes())
	return err
}
