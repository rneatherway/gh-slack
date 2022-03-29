package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"regexp"
	"rneatherway/slack-to-md/slackclient"
	"strconv"
)

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

func realMain() error {
	channelID, timestamp, err := parsePermalink(*startFlag)
	if err != nil {
		return err
	}

	limit, err := strconv.Atoi(*endFlag)
	if err != nil {
		return err
	}

	logger := log.New(io.Discard, "", log.LstdFlags)
	if *verboseFlag {
		logger = log.Default()
	}

	client, err := slackclient.NewSlackClient(
		"github",
		logger)
	if err != nil {
		return err
	}

	history, err := client.History(channelID, timestamp, limit)
	if err != nil {
		return err
	}

	markdown, err := convertMessagesToMarkdown(client, history.Messages)
	if err != nil {
		return err
	}

	fmt.Println(markdown)

	return nil
}

// TODO: allow grabbing a thread
// TODO: try a permalink-based interface. E.g:
//   ./slack-to-md https://github.slack.com/archives/C0H15BV4K/p1648489045530009 -- get a thread
//   ./slack-to-md https://github.slack.com/archives/C0H15BV4K/p1648489045530009 30 -- get 30 messages starting at that one
//   ./slack-to-md \
//       https://github.slack.com/archives/C0H15BV4K/p1648489045530009 \
//       https://github.slack.com/archives/C0H15BV4K/p1648489045530009 -- get messages between these two, inclusive
var startFlag = flag.String("start", "", "Permalink to start reading messages from. If a thread, that entire thread will be read.")
var endFlag = flag.String("end", "20", "Permalink of last message to read, or an integer number of messages to read.")
var verboseFlag = flag.Bool("v", false, "Verbose output")

func main() {
	flag.Parse()
	err := realMain()
	if err != nil {
		fmt.Println(err)
	}
}
