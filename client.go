package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

type Cursor struct {
	NextCursor string `json:"next_cursor"`
}

type CursorResponseMetadata struct {
	ResponseMetadata Cursor `json:"response_metadata"`
}

type Message struct {
	User     string
	Username string
	Text     string
	Ts       string
}

type HistoryResponse struct {
	CursorResponseMetadata
	Ok       bool
	Messages []Message
}

type Channel struct {
	ID         string
	Name       string
	Is_Channel bool
}

type ConversationsResponse struct {
	CursorResponseMetadata
	Ok       bool
	Channels []Channel
}

type User struct {
	ID   string
	Name string
}

type UsersResponse struct {
	Ok      bool
	Members []User
}

type Cache struct {
	Channels map[string]string
	Users    map[string]string
}

type SlackClient struct {
	cachePath string
	team      string
	client    http.Client
	auth      *SlackAuth
	cache     Cache
	log       *log.Logger
}

func NewSlackClient(cachePath string, auth *SlackAuth, team string, log *log.Logger) (*SlackClient, error) {
	client := &SlackClient{
		cachePath: cachePath,
		auth:      auth,
		team:      team,
		log:       log,
	}
	// TODO: this isn't safe, either move SlackClient to another package so that you have to use this constructor
	// or all the 'public' methods have to guarantee that it's loaded.
	err := client.loadCache()
	return client, err
}

// TODO: change client receiver to c
func (client *SlackClient) get(path string, params map[string]string) ([]byte, error) {
	u, err := url.Parse(fmt.Sprintf("https://%s.slack.com/api/", client.team))
	if err != nil {
		return nil, err
	}
	u.Path += path
	q := u.Query()
	q.Add("token", client.auth.Token)
	for p := range params {
		q.Add(p, params[p])
	}
	u.RawQuery = q.Encode()

	var body []byte
	for {
		req, err := http.NewRequest("GET", u.String(), nil)
		if err != nil {
			return nil, err
		}
		for key := range client.auth.Cookies {
			req.AddCookie(&http.Cookie{Name: key, Value: client.auth.Cookies[key]})
		}

		resp, err := client.client.Do(req)
		if err != nil {
			return nil, err
		}

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == 429 {
			s, err := strconv.Atoi(resp.Header["Retry-After"][0])
			if err != nil {
				return nil, err
			}
			d := time.Duration(s)
			client.log.Printf("rate limited, waiting %ds", d)
			time.Sleep(d * time.Second)
		} else if resp.StatusCode >= 300 {
			return nil, fmt.Errorf("status code %d, headers: %q, body: %q", resp.StatusCode, resp.Header, body)
		} else {
			break
		}
	}

	return body, nil
}

func (client *SlackClient) conversations(params map[string]string) ([]Channel, error) {
	channels := make([]Channel, 0, 1000)
	conversations := &ConversationsResponse{}
	for {
		client.log.Printf("Fetching conversations with cursor %q", conversations.ResponseMetadata.NextCursor)
		body, err := client.get("conversations.list",
			map[string]string{
				"cursor":           conversations.ResponseMetadata.NextCursor,
				"exclude_archived": "true"},
		)
		if err != nil {
			return nil, err
		}

		if err = json.Unmarshal(body, conversations); err != nil {
			return nil, err
		}

		if !conversations.Ok {
			return nil, fmt.Errorf("conversations response not OK: %s", body)
		}

		channels = append(channels, conversations.Channels...)
		client.log.Printf("Fetched %d channels (total so far %d)",
			len(conversations.Channels),
			len(channels))

		if conversations.ResponseMetadata.NextCursor == "" {
			break
		}
	}

	return channels, nil
}

func (client *SlackClient) users(params map[string]string) (*UsersResponse, error) {
	body, err := client.get("users.list", nil)
	if err != nil {
		return nil, err
	}

	users := &UsersResponse{}
	err = json.Unmarshal(body, users)
	if err != nil {
		return nil, err
	}

	if !users.Ok {
		return nil, fmt.Errorf("users response not OK: %s", body)
	}

	return users, nil
}

func (client *SlackClient) loadCache() error {
	content, err := os.ReadFile(client.cachePath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}

	return json.Unmarshal(content, &client.cache)
}

func (client *SlackClient) history(channelID string, startTime time.Time, endTime time.Time) (*HistoryResponse, error) {
	body, err := client.get("conversations.history",
		map[string]string{
			"channel": channelID,
			"oldest":  strconv.FormatInt(startTime.Unix(), 10),
			"latest":  strconv.FormatInt(endTime.Unix(), 10)})
	if err != nil {
		return nil, err
	}

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

func (client *SlackClient) saveCache() error {
	bs, err := json.Marshal(client.cache)
	if err != nil {
		return err
	}

	err = os.WriteFile(client.cachePath, bs, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (client *SlackClient) getChannelID(name string) (string, error) {
	if id, ok := client.cache.Channels[name]; ok {
		return id, nil
	}

	channels, err := client.conversations(nil)
	if err != nil {
		return "", err
	}

	client.cache.Channels = make(map[string]string)
	for _, ch := range channels {
		client.cache.Channels[ch.Name] = ch.ID
	}

	err = client.saveCache()
	if err != nil {
		return "", err
	}

	if id, ok := client.cache.Channels[name]; ok {
		return id, nil
	}

	return "", fmt.Errorf("no channel with name %q", name)
}

func (client *SlackClient) getUsername(id string) (string, error) {
	if id, ok := client.cache.Users[id]; ok {
		return id, nil
	}

	ur, err := client.users(nil)
	if err != nil {
		return "", err
	}

	client.cache.Users = make(map[string]string)
	for _, ch := range ur.Members {
		client.cache.Users[ch.ID] = ch.Name
	}

	err = client.saveCache()
	if err != nil {
		return "", err
	}

	if id, ok := client.cache.Users[id]; ok {
		return id, nil
	}

	return "", fmt.Errorf("no user with id %q", id)
}
