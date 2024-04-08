package testing

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"maps"
	"net/http"
	"os"
	"slices"
	"strings"
)

type Replayer struct {
	exchanges []Exchange
	matcher   func(*http.Request, Request) bool
}

func NewReplayer(file string) (*Replayer, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var exchanges []Exchange
	err = json.NewDecoder(f).Decode(&exchanges)
	if err != nil {
		return nil, err
	}

	return &Replayer{
		exchanges: exchanges,
		matcher:   DefaultMatcher,
	}, nil
}

var _ http.RoundTripper = (*Replayer)(nil)

func (r *Replayer) RoundTrip(req *http.Request) (*http.Response, error) {
	if len(r.exchanges) == 0 {
		return nil, errors.New("no exchanges remain to replay")
	}

	// TODO: we just take the first one. should we rather search for a match? Check ruby VCR featureset for inspiration
	ex := r.exchanges[0]
	r.exchanges = r.exchanges[1:]

	// query parameters
	// allow debugging lack of match
	if r.matcher(req, ex.Request) {
		var r io.Reader
		if string(ex.Response.BodyBytes) != "" {
			r = bytes.NewReader(ex.Response.BodyBytes)
		} else {
			r = strings.NewReader(ex.Response.BodyString)
		}
		return &http.Response{
			Body:       io.NopCloser(r),
			StatusCode: ex.Response.StatusCode,
		}, nil
	}

	return nil, errors.New("no match")
}

func DefaultMatcher(req *http.Request, stored Request) bool {
	return req.Method == stored.Method && req.URL.String() == stored.URL &&
		maps.EqualFunc(req.Header, stored.Headers, slices.Equal)
}
