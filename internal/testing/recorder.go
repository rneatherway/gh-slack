package testing

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"unicode/utf8"
)

type Recorder struct {
	file  string
	inner http.RoundTripper

	exchanges []Exchange
}

var _ http.RoundTripper = (*Recorder)(nil)

func NewRecorder(file string, inner http.RoundTripper) *Recorder {
	return &Recorder{file: file, inner: inner}
}

func (r *Recorder) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := r.inner.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	ex := Exchange{
		Request: Request{
			URL:     req.URL.String(),
			Headers: req.Header.Clone(),
			Method:  req.Method,
		},
		Response: Response{
			Headers:    resp.Header,
			StatusCode: resp.StatusCode,
		},
	}

	if utf8.Valid(body) {
		ex.Response.BodyString = string(body)
	} else {
		ex.Response.BodyBytes = body
	}

	// TODO: what is actually happening with this slice if it's a pointer vs a value?
	r.exchanges = append(r.exchanges, ex)

	// Put the Body back in the response for the actual client code to read.
	resp.Body = io.NopCloser(bytes.NewReader(body))

	return resp, err
}

func (r *Recorder) Close() error {
	f, err := os.Create(r.file)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(r.exchanges)
}
