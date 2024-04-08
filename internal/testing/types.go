package testing

import "net/http"

type Exchange struct {
	Request  Request  `json:"request"`
	Response Response `json:"response"`
}

type Request struct {
	URL     string      `json:"url"`
	Headers http.Header `json:"headers"`
	Method  string      `json:"method"`

	// Body is deliberately excluded so far, for simplicity.
}

type Response struct {
	Headers    http.Header `json:"headers"`
	BodyString string      `json:"body_string,omitempty"`
	BodyBytes  []byte      `json:"body_bytes,omitempty"`
	StatusCode int         `json:"status_code"`
}
