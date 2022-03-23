package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Message struct {
	User     string
	Username string
	Text     string
	Ts       string
}

type HistoryResponse struct {
	Ok       bool
	Messages []Message
}

type Channel struct {
	ID         string
	Name       string
	Is_Channel bool
}

type ListResponse struct {
	Ok       bool
	Channels []Channel
}

func mkRequest(path string, params map[string]string) ([]byte, error) {
	// TODO: Read this from the environment or something
	token := ""
	cookies := map[string]string{
		"d": "",
	}

	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Add("token", token)
	for p := range params {
		q.Add(p, params[p])
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	for key := range cookies {
		req.AddCookie(&http.Cookie{Name: key, Value: cookies[key]})
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func getChannelID(name string) (string, error) {
	// TODO: this needs to use paging, and also to cache the channel name to id mappings
	body, err := mkRequest("https://github.slack.com/api/conversations.list", map[string]string{"limit": "5"})
	if err != nil {
		return "", err
	}

	listResponse := &ListResponse{}
	err = json.Unmarshal(body, listResponse)
	if err != nil {
		return "", err
	}

	if !listResponse.Ok {
		return "", fmt.Errorf("list response not OK: %s", body)
	}

	for _, channel := range listResponse.Channels {
		if channel.Name == name {
			return channel.ID, nil
		}
	}

	return "", fmt.Errorf("channel with name '#%s' not found", name)
}

func getConversationHistory(channelID string, startTime time.Time, endTime time.Time) (*HistoryResponse, error) {
	// TODO: long-term this will need paging
	body, err := mkRequest("https://github.slack.com/api/conversations.history",
		map[string]string{"channel": channelID, "oldest": strconv.FormatInt(startTime.Unix(), 10), "latest": strconv.FormatInt(endTime.Unix(), 10)})
	if err != nil {
		return nil, err
	}

	// TODO: verbose flag?
	// out := &bytes.Buffer{}
	// err = json.Indent(out, body, "", "  ")
	// if err != nil {
	// 	return nil, err
	// }
	// fmt.Println(out.String())

	historyResponse := &HistoryResponse{}
	err = json.Unmarshal(body, historyResponse)
	if err != nil {
		return nil, err
	}

	if !historyResponse.Ok {
		return nil, fmt.Errorf("history response not OK: %s", body)
	}

	return historyResponse, nil
}

func convertMessagesToMarkdown(messages []Message) (string, error) {
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

		// TODO: need a user cache
		// TODO: convert @mentions to be backticked so as not to ping people
		b.WriteString(fmt.Sprintf("> **%s** at %s\n> %s\n", message.User, tm.Format(time.RFC3339), message.Text))
	}

	return b.String(), nil
}

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

	channelID, err := getChannelID(*channelFlag)
	if err != nil {
		return err
	}

	history, err := getConversationHistory(channelID, startTime, endTime)
	if err != nil {
		return err
	}

	markdown, err := convertMessagesToMarkdown(history.Messages)
	if err != nil {
		return err
	}

	fmt.Println(markdown)

	return nil
}

var channelFlag = flag.String("channel", "", "Channel name to read from")
var startFlag = flag.String("start", "", "Retrieve messages after this time")
var endFlag = flag.String("end", "", "Retrieve messages before this time")

func main() {
	flag.Parse()
	err := realMain()
	if err != nil {
		fmt.Println(err)
	}
}
