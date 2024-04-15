package markdown

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/rneatherway/gh-slack/internal/slackclient"
	"github.com/rneatherway/slack/pkg/markdown"
)

func convert(client *slackclient.SlackClient, b *strings.Builder, s string) error {
	text, err := markdown.Convert(client, s)
	if err != nil {
		return err
	}

	for _, line := range strings.Split(text, "\n") {
		// TODO: Might be a good idea to escape 'line'
		fmt.Fprintf(b, "> %s\n", line)
	}

	return nil
}

func FromMessages(client *slackclient.SlackClient, history *slackclient.HistoryResponse) (string, error) {
	b := &strings.Builder{}
	messages := history.Messages
	msgTimes := make(map[string]time.Time, len(messages))

	for _, message := range messages {
		tm, err := markdown.ParseUnixTimestamp(message.Ts)
		if err != nil {
			return "", err
		}

		msgTimes[message.Ts] = *tm
	}

	// It's surprising that these messages are not already always returned in date order,
	// and actually I observed initially that they seemed to be, but at least some of the
	// time they are returned in reverse order so it's simpler to just sort them now.
	sort.Slice(messages, func(i, j int) bool {
		return msgTimes[messages[i].Ts].Before(msgTimes[messages[j].Ts])
	})

	lastSpeakerID := ""

	for i, message := range messages {
		username, err := client.UsernameForMessage(message)
		if err != nil {
			return "", err
		}

		speakerID := message.User
		if speakerID == "" {
			speakerID = message.BotID
		}

		messageTime := msgTimes[message.Ts]
		messageTimeDiffInMinutes := 0

		// How far apart in minutes can two messages be, by the same author, before we repeat the header line?
		messageTimeMinuteCutoff := 60

		if i > 0 {
			prevMessage := messages[i-1]
			prevMessageTime := msgTimes[prevMessage.Ts]
			messageTimeDiffInMinutes = int(messageTime.Sub(prevMessageTime).Minutes())
		}

		if lastSpeakerID != "" && speakerID != lastSpeakerID || messageTimeDiffInMinutes > messageTimeMinuteCutoff {
			fmt.Fprintf(b, "\n")
		}

		includeSpeakerHeader := lastSpeakerID == "" || speakerID != lastSpeakerID ||
			messageTimeDiffInMinutes > messageTimeMinuteCutoff

		if includeSpeakerHeader {
			fmt.Fprintf(b, "> **%s** at %s\n",
				username,
				messageTime.In(client.GetLocation()).Format("2006-01-02 15:04 MST"))
		}
		fmt.Fprintf(b, ">\n")

		if message.Text != "" {
			err = convert(client, b, message.Text)
			if err != nil {
				return "", err
			}
		}

		// These seem to be mostly bot messages so far. Perhaps we should just skip them?
		for _, a := range message.Attachments {
			err = convert(client, b, a.Text)
			if err != nil {
				return "", err
			}
		}

		if !includeSpeakerHeader {
			b.WriteString("\n")
		}

		lastSpeakerID = speakerID
	}

	return b.String(), nil
}

func WrapInDetails(channelName, link, s string) string {
	return fmt.Sprintf("Slack conversation archive of [`#%s`](%s)\n\n<details>\n  <summary>Click to expand</summary>\n\n%s\n</details>",
		channelName, link, s)
}
