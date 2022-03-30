package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"rneatherway/slack-to-md/slackclient"

	"github.com/jessevdk/go-flags"
)

// TODO: Move into `internal` directory so we don't get imported

var permalinkRE = regexp.MustCompile("https://[^./]+.slack.com/archives/([A-Z0-9]+)/p([0-9]+)([0-9]{6})")

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
		Start string `description:"Permalink for the first message to fetch. Following messages are then fetched from that channel (or thread if applicable)" required:"true"`
	} `positional-args:"yes"`
	Limit   int  `short:"l" long:"limit" default:"20" description:"Number of _channel_ messages to be fetched after the starting message (all thread messages are fetched)"`
	Verbose bool `short:"v" long:"verbose" description:"Show verbose debug information"`
}

func realMain() error {
	_, err := flags.NewParser(&opts, flags.HelpFlag|flags.PassDoubleDash).Parse()
	if err != nil {
		return err
	}

	channelID, timestamp, err := parsePermalink(opts.Args.Start)
	if err != nil {
		return err
	}

	logger := log.New(io.Discard, "", log.LstdFlags)
	if opts.Verbose {
		logger = log.Default()
	}

	client, err := slackclient.NewSlackClient(
		"github",
		logger)
	if err != nil {
		return err
	}

	history, err := client.History(channelID, timestamp, opts.Limit)
	if err != nil {
		return err
	}

	markdown, err := convertMessagesToMarkdown(client, history)
	if err != nil {
		return err
	}

	fmt.Println(markdown)

	return nil
}

func main() {
	err := realMain()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
