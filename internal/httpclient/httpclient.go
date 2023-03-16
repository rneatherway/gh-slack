package httpclient

import (
	"net/http"
)

// See https://www.thegreatcodeadventure.com/mocking-http-requests-in-golang/
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var (
	Client HTTPClient
)

func init() {
	Client = &http.Client{}
}
