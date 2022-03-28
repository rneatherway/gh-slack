package main

import (
	"encoding/json"
	"fmt"
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

func TestUserRE(t *testing.T) {
	results := userRE.FindAllStringIndex("<@UP7UAV3NH> <@UPA5ANVNJ> hello", -1)
	if len(results) != 2 {
		t.Errorf("results length %d", len(results))
	}
	if !(results[0][0] == 0 && results[0][1] == 12) {
		t.Errorf("first match %v", results[0])
	}
	if !(results[1][0] == 13 && results[1][1] == 25) {
		t.Errorf("second match %v", results[1])
	}
}

type TestUserProvider struct {
	counter int
}

func (c *TestUserProvider) getUsername(id string) (string, error) {
	c.counter += 1
	return fmt.Sprintf("test_username_%d", c.counter), nil
}

func TestInterpolateUsers(t *testing.T) {
	table := [][]string{
		{"<@UP7UAV3NH>", "`@test_username_1`"},
		{"<@UP7UAV3NH> hi hi", "`@test_username_1` hi hi"},
		{"hi<@UP7UAV3NH> hi hi", "hi`@test_username_1` hi hi"},
		{"<@UP7UAV3NH> hello <@UP756V3NH>", "`@test_username_1` hello `@test_username_2`"},
		{"<@UP7UAV3NH> <@UP756V3NH> hello", "`@test_username_1` `@test_username_2` hello"},
	}

	for _, test := range table {
		input := test[0]
		expected := test[1]
		actual, _ := interpolateUsers(&TestUserProvider{}, input)

		if actual != expected {
			t.Errorf("expected %q, actual %q", expected, actual)
		}
	}
}
