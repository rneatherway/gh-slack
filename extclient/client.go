package extclient

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/rneatherway/gh-slack/internal/httpclient"
)

type SlackClient struct {
	team string
	auth *SlackAuth
}

func New(team string) (*SlackClient, error) {
	auth, err := getSlackAuth(team)
	if err != nil {
		return nil, err
	}

	c := &SlackClient{
		team: team,
		auth: auth,
	}

	return c, nil
}

func (c *SlackClient) API(verb, path string, params map[string]string, body []byte) ([]byte, error) {
	u, err := url.Parse(fmt.Sprintf("https://%s.slack.com/api/", c.team))
	if err != nil {
		return nil, err
	}
	u.Path += path
	q := u.Query()
	for p := range params {
		q.Add(p, params[p])
	}
	u.RawQuery = q.Encode()

	reqBody := bytes.NewReader(body)
	var resBody []byte

	for {
		req, err := http.NewRequest(verb, u.String(), reqBody)
		if err != nil {
			return nil, err
		}
		// FIXME: this doesn't seem to break non-POST/non-data requests, but might
		// be polluting the headers.
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.auth.Token))
		for key := range c.auth.Cookies {
			req.AddCookie(&http.Cookie{Name: key, Value: c.auth.Cookies[key]})
		}

		resp, err := httpclient.Client.Do(req)
		if err != nil {
			return nil, err
		}

		resBody, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == 429 {
			s, err := strconv.Atoi(resp.Header["Retry-After"][0])
			if err != nil {
				return nil, err
			}
			d := time.Duration(s)
			time.Sleep(d * time.Second)
		} else if resp.StatusCode >= 300 {
			return nil, fmt.Errorf("status code %d, headers: %q, body: %q", resp.StatusCode, resp.Header, body)
		} else {
			break
		}
	}

	return resBody, nil
}
