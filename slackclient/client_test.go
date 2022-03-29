package slackclient

import (
	"encoding/json"
	"testing"
)

type Y struct {
	NextCursor string `json:"next_cursor"`
}

type X struct {
	ResponseMetadata Y `json:"response_metadata"`
}

func TestUnmarshaling(t *testing.T) {
	input := []byte("{\"response_metadata\":{\"next_cursor\":\"dGVhbTpDMEY3SEpCNlk=\"}}")
	bs := &X{}
	err := json.Unmarshal(input, bs)

	if err != nil {
		t.Error(err)
	}

	if bs.ResponseMetadata.NextCursor == "" {
		t.Error()
	}
}
