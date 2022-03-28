package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"time"
)

func realMain() error {
	if *channelFlag == "" {
		return errors.New("channel name is required")
	}
	if *startFlag == "" {
		return errors.New("start time is required")
	}
	if *endFlag == "" {
		return errors.New("end time is required")
	}

	startTime, err := time.Parse("2006-01-02 15:04", *startFlag)
	if err != nil {
		return errors.New("start time not in expected format")
	}

	endTime, err := time.Parse("2006-01-02 15:04", *endFlag)
	if err != nil {
		return errors.New("start time not in expected format")
	}

	logger := log.New(io.Discard, "", log.LstdFlags)
	if *verboseFlag {
		logger = log.Default()
	}

	// TODO: should this be part of the SlackClient?
	auth, err := getSlackAuth("github")
	if err != nil {
		return err
	}

	client, err := NewSlackClient(
		"cache.json", // TODO: Move this to XDG_DATA_HOME or ~/.local/share (or MacOS equivalent?)
		auth,
		"github",
		logger)
	if err != nil {
		return err
	}

	channelID, err := client.getChannelID(*channelFlag)
	if err != nil {
		return err
	}

	history, err := client.history(channelID, startTime, endTime)
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
var channelFlag = flag.String("channel", "", "Channel name to read from")
var startFlag = flag.String("start", "", "Retrieve messages after this time")
var endFlag = flag.String("end", "", "Retrieve messages before this time")
var verboseFlag = flag.Bool("v", false, "Verbose output")

func main() {
	log.Default()
	flag.Parse()
	err := realMain()
	if err != nil {
		fmt.Println(err)
	}
}
