package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var userRE = regexp.MustCompile("<@[A-Z0-9]+>")

type UserProvider interface {
	getUsername(string) (string, error)
}

func interpolateUsers(client UserProvider, s string) (string, error) {
	userLocations := userRE.FindAllStringIndex(s, -1)
	out := &strings.Builder{}
	last := 0
	for _, userLocation := range userLocations {
		start := userLocation[0]
		end := userLocation[1]

		username, err := client.getUsername(s[start+2 : end-1])
		if err != nil {
			return "", err
		}
		out.WriteString(s[last:start])
		out.WriteString("`@")
		out.WriteString(username)
		out.WriteRune('`')
		last = end
	}
	out.WriteString(s[last:])

	return out.String(), nil
}

func convertMessagesToMarkdown(client *SlackClient, messages []Message) (string, error) {
	b := &strings.Builder{}

	for _, message := range messages {
		tsParts := strings.Split(message.Ts, ".")
		if len(tsParts) != 2 {
			return "", fmt.Errorf("timestamp '%s' in not in <seconds>.<milliseconds> format", message.Ts)
		}

		msgTime, err := strconv.ParseInt(tsParts[0], 10, 64)
		if err != nil {
			return "", err
		}

		tm := time.Unix(msgTime, 0)

		username, err := client.getUsername(message.User)
		if err != nil {
			return "", err
		}

		b.WriteString("> **")
		b.WriteString(username)
		b.WriteString("** at ")
		b.WriteString(tm.Format("2006-01-02 15:04"))
		b.WriteString("\n>\n")
		text, err := interpolateUsers(client, message.Text)
		if err != nil {
			return "", err
		}

		for _, line := range strings.Split(text, "\n") {
			b.WriteString("> ")
			b.WriteString(line)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	return b.String(), nil
}
