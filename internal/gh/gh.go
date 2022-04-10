package gh

import (
	"fmt"
	"os"

	"github.com/cli/go-gh"
	"github.com/rneatherway/gh-slack/internal/slackclient"
)

func NewIssue(repoUrl string, channel *slackclient.Channel, content string) error {
	out, _, err := gh.Exec(
		"issue",
		"-R",
		repoUrl,
		"create",
		"--title",
		fmt.Sprintf("Slack conversation archive of `#%s`", channel.Name),
		"--body",
		content)
	os.Stdout.Write(out.Bytes())
	return err
}

func AddComment(issueUrl string, channel *slackclient.Channel, content string) error {
	out, _, err := gh.Exec(
		"issue",
		"comment",
		issueUrl,
		"--body",
		fmt.Sprintf("Slack conversation archive of `#%s`\n%s", channel.Name, content))
	os.Stdout.Write(out.Bytes())
	return err
}
