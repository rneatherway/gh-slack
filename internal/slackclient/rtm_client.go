package slackclient

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cli/go-gh/pkg/markdown"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type RTMClient struct {
	conn *websocket.Conn
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
		if message.Channel == channelID && message.Type == "message" && strings.EqualFold(message.BotProfile.Name, botName) {
			for _, attachment := range message.Attachments {
				s, err := markdown.Render(attachment.Text)
				if err != nil {
					return err
				}
				s = strings.TrimRight(s, " \t\n")
				fmt.Println(s)
			}
			break
		}
	}
	return nil
}

func (c *RTMClient) Close() error {
	return c.conn.Close(websocket.StatusNormalClosure, "")
}
