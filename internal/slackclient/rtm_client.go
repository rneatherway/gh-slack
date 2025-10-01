package slackclient

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cli/go-gh/v2/pkg/markdown"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type RTMClient struct {
	conn        *websocket.Conn
	slackClient *SlackClient
}

type File struct {
	Preview string `json:"preview"`
}

type RTMEvent struct {
	Type        string       `json:"type"`
	Channel     string       `json:"channel,omitempty"`
	User        string       `json:"user,omitempty"`
	Text        string       `json:"text,omitempty"`
	TS          string       `json:"ts,omitempty"`
	BotID       string       `json:"bot_id,omitempty"`
	BotProfile  BotProfile   `json:"bot_profile,omitempty"`
	Subtype     string       `json:"subtype,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
	Files       []File       `json:"files,omitempty"`
}

// correctUser checks if the message is sent by the bot/user that we are waiting
// for. We accept three possible matches against the user-provided name:
//   - The bot profile's name (case-insensitive)
//   - The user's ID (case-sensitive)
//   - The user's name (case-insensitive)
func (c *RTMClient) correctUser(message *RTMEvent, botName string) bool {
	if strings.EqualFold(message.BotProfile.Name, botName) {
		return true
	}

	if message.User == botName {
		return true
	}

	// It would be nice to just convert botName to an ID and compare that, but
	// the Slack API doesn't provide a way to do that if botName is not a member
	// of the team (an outside collaborator). So we have to do this the hard
	// way.
	user, err := c.slackClient.UsernameForID(message.User)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return false
	}

	return strings.EqualFold(user, botName)
}

func trimAndPrint(text string) {
	s, err := markdown.Render(text)
	if err != nil {
		// This is a bit lazy, but the default configuration of the markdown
		// renderer cannot fail.
		panic(err)
	}

	s = strings.TrimRight(s, " \t\n")
	if s == "" {
		return
	}

	fmt.Printf("%s\n", s)
}

// ListenForMessagesFromBot listens for the first message from the bot in a given channel and prints its contents
func (c *RTMClient) ListenForMessagesFromBot(channelID, botName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	for {
		message := &RTMEvent{}
		err := wsjson.Read(ctx, c.conn, &message)
		if err != nil {
			c.conn.Close(websocket.StatusUnsupportedData, "")
			return err
		}

		if message.Channel == channelID && message.Type == "message" && c.correctUser(message, botName) {
			trimAndPrint(message.Text)

			for _, attachment := range message.Attachments {
				trimAndPrint(attachment.Text)
			}

			for _, file := range message.Files {
				trimAndPrint(file.Preview)
			}

			break
		}
	}
	return nil
}

func (c *RTMClient) Close() error {
	return c.conn.Close(websocket.StatusNormalClosure, "")
}
