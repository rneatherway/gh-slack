package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/rneatherway/gh-slack/internal/gh"
	"github.com/rneatherway/gh-slack/internal/markdown"
	"github.com/rneatherway/gh-slack/internal/slackclient"
	"github.com/rneatherway/gh-slack/internal/version"

	"github.com/jessevdk/go-flags"

	github "github.com/cli/go-gh"
)

var (
	permalinkRE = regexp.MustCompile("https://[^./]+.slack.com/archives/([A-Z0-9]+)/p([0-9]+)([0-9]{6})")
	nwoRE       = regexp.MustCompile("^/[^/]+/[^/]+/?$")
	issueRE     = regexp.MustCompile("^/[^/]+/[^/]+/issues/[0-9]+/?$")
)

// https://github.slack.com/archives/CP9GMKJCE/p1648028606962719
// returns (CP9GMKJCE, 1648028606.962719, nil)
func parsePermalink(link string) (string, string, error) {
	result := permalinkRE.FindStringSubmatch(link)
	if result == nil {
		return "", "", fmt.Errorf("not a permalink: %q", link)
	}

	return result[1], result[2] + "." + result[3], nil
}

var opts struct {
	Args struct {
		Start string `description:"Required. Permalink for the first message to fetch. Following messages are then fetched from that channel (or thread if applicable)"`
	} `positional-args:"yes"`
	Limit   int    `short:"l" long:"limit" default:"20" description:"Number of _channel_ messages to be fetched after the starting message (all thread messages are fetched)"`
	Verbose bool   `short:"v" long:"verbose" description:"Show verbose debug information"`
	Version bool   `long:"version" description:"Output version information"`
	Details bool   `short:"d" long:"details" description:"Wrap the markdown output in HTML <details> tags"`
	Issue   string `short:"i" long:"issue" description:"The URL of a repository to post the output as a new issue, or the URL of an issue to add a comment to that issue"`
	Team    string `short:"t" long:"team" description:"The slack team name of the workspace you want to use"`
}

func realMain() error {
	_, err := flags.NewParser(&opts, flags.HelpFlag|flags.PassDoubleDash).Parse()
	if err != nil {
		return err
	}

	if opts.Version {
		fmt.Printf("gh-slack %s (%s)\n", version.Version(), version.Commit())
		return nil
	}

	if opts.Args.Start == "" {
		return errors.New("the required argument `Start` was not provided")
	}

	var repoUrl, issueUrl string
	if opts.Issue != "" {
		u, err := url.Parse(opts.Issue)
		if err != nil {
			return err
		}

		if nwoRE.MatchString(u.Path) {
			repoUrl = opts.Issue
		} else if issueRE.MatchString(u.Path) {
			issueUrl = opts.Issue
		} else {
			return fmt.Errorf("not a repository or issue URL: %q", opts.Issue)
		}
	}

	channelID, timestamp, err := parsePermalink(opts.Args.Start)
	if err != nil {
		return err
	}

	logger := log.New(io.Discard, "", log.LstdFlags)
	if opts.Verbose {
		logger = log.Default()
	}

	var team string

	if opts.Team == "" {
		team, _, err = getSlackTeam()
		if err != nil {
			return errors.New("No team set. Either pass -t/--team or run `gh config set extensions.slack.team team`.")
		}
	} else {
		team = opts.Team
	}

	client, err := slackclient.New(
		team,
		logger)
	if err != nil {
		return err
	}

	history, err := client.History(channelID, timestamp, opts.Limit)
	if err != nil {
		return err
	}

	output, err := markdown.FromMessages(client, history)
	if err != nil {
		return err
	}

	var channelName string
	if opts.Details {
		channelInfo, err := client.ChannelInfo(channelID)
		if err != nil {
			return err
		}

		channelName = channelInfo.Name
		output = markdown.WrapInDetails(channelName, opts.Args.Start, output)
	}

	if repoUrl != "" {
		if channelName == "" {
			channelInfo, err := client.ChannelInfo(channelID)
			if err != nil {
				return err
			}
			channelName = channelInfo.Name
		}

		err := gh.NewIssue(repoUrl, channelName, output)
		if err != nil {
			return err
		}
	} else if issueUrl != "" {
		err := gh.AddComment(issueUrl, output)
		if err != nil {
			return err
		}
	} else {
		os.Stdout.WriteString(output)
	}

	return nil
}

func main() {
	err := realMain()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func getSlackTeam() (string, string, error) {
	out, stderr, err := github.Exec(
		"config",
		"get",
		"extensions.slack.team",
	)
	return strings.TrimRight(string(out.Bytes()), "\n"), string(stderr.Bytes()), err
}
