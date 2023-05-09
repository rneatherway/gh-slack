package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"regexp"

	"github.com/rneatherway/gh-slack/internal/gh"
	"github.com/rneatherway/gh-slack/internal/markdown"
	"github.com/rneatherway/gh-slack/internal/slackclient"
	"github.com/rneatherway/gh-slack/internal/version"

	"github.com/jessevdk/go-flags"
)

var (
	permalinkRE = regexp.MustCompile("https://([^./]+).slack.com/archives/([A-Z0-9]+)/p([0-9]+)([0-9]{6})")
	nwoRE       = regexp.MustCompile("^/[^/]+/[^/]+/?$")
	issueRE     = regexp.MustCompile("^/[^/]+/[^/]+/(issues|pull)/[0-9]+/?$")
)

type linkParts struct {
	team      string
	channelID string
	timestamp string
}

// https://github.slack.com/archives/CP9GMKJCE/p1648028606962719
// returns (github, CP9GMKJCE, 1648028606.962719, nil)
func parsePermalink(link string) (linkParts, error) {
	result := permalinkRE.FindStringSubmatch(link)
	if result == nil {
		return linkParts{}, fmt.Errorf("not a permalink: %q", link)
	}

	return linkParts{
		team:      result[1],
		channelID: result[2],
		timestamp: result[3] + "." + result[4],
	}, nil
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

	var repoUrl, issueOrPrUrl, subCmd string
	if opts.Issue != "" {
		u, err := url.Parse(opts.Issue)
		if err != nil {
			return err
		}

		matches := issueRE.FindStringSubmatch(u.Path)
		if matches != nil {
			issueOrPrUrl = opts.Issue
			subCmd = "issue"
			if matches[1] == "pull" {
				subCmd = "pr"
			}
		} else if nwoRE.MatchString(u.Path) {
			repoUrl = opts.Issue
		} else {
			return fmt.Errorf("not a repository or issue URL: %q", opts.Issue)
		}
	}

	linkParts, err := parsePermalink(opts.Args.Start)
	if err != nil {
		return err
	}

	logger := log.New(io.Discard, "", log.LstdFlags)
	if opts.Verbose {
		logger = log.Default()
	}

	client, err := slackclient.New(
		linkParts.team,
		logger)
	if err != nil {
		return err
	}

	history, err := client.History(linkParts.channelID, linkParts.timestamp, opts.Limit)
	if err != nil {
		return err
	}

	output, err := markdown.FromMessages(client, history)
	if err != nil {
		return err
	}

	var channelName string
	if opts.Details {
		channelInfo, err := client.ChannelInfo(linkParts.channelID)
		if err != nil {
			return err
		}

		channelName = channelInfo.Name
		output = markdown.WrapInDetails(channelName, opts.Args.Start, output)
	}

	if repoUrl != "" {
		if channelName == "" {
			channelInfo, err := client.ChannelInfo(linkParts.channelID)
			if err != nil {
				return err
			}
			channelName = channelInfo.Name
		}

		err := gh.NewIssue(repoUrl, channelName, output)
		if err != nil {
			return err
		}
	} else if issueOrPrUrl != "" {
		err := gh.AddComment(subCmd, issueOrPrUrl, output)
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
