package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Message struct {
	User string
	Text string
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

func mkRequest(path string) ([]byte, error) {
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
	q.Add("limit", "5")
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

func realMain() error {
	body, err := mkRequest("https://github.slack.com/api/conversations.list")
	if err != nil {
		return err
	}

	listResponse := &ListResponse{}
	err = json.Unmarshal(body, listResponse)
	if err != nil {
		return err
	}

	if !listResponse.Ok {
		return errors.New("list response not OK")
	}

	for _, channel := range listResponse.Channels {
		fmt.Printf("%s: %s (%t)\n", channel.ID, channel.Name, channel.Is_Channel)
	}

	body, err = mkRequest(fmt.Sprintf("https://github.slack.com/api/conversations.history?channel=%s", listResponse.Channels[0].ID))
	if err != nil {
		return err
	}

	historyResponse := &HistoryResponse{}
	err = json.Unmarshal(body, historyResponse)
	if err != nil {
		return err
	}

	if !historyResponse.Ok {
		fmt.Println(string(body))
		return errors.New("history response not OK")
	}

	for _, message := range historyResponse.Messages {
		fmt.Printf("%s: %s\n", message.User, message.Text)
	}

	return nil
}

func main() {
	err := realMain()
	if err != nil {
		fmt.Println(err)
	}
}
