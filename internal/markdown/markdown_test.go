package markdown

import (
	"fmt"
	"testing"
)

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

func (c *TestUserProvider) UsernameForID(id string) (string, error) {
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

func TestLinkify(t *testing.T) {
	table := [][]string{
		{"Hello <https://example.com|text> end", "Hello [text](https://example.com) end"},
		{"Hello <https://example.com|text> end and here is another link <http://github.com|GitHub!!> go check it out", "Hello [text](https://example.com) end and here is another link [GitHub!!](http://github.com) go check it out"},
	}

	for _, test := range table {
		input := test[0]
		expected := test[1]
		actual := linkRE.ReplaceAllString(input, "[$2]($1)")

		if actual != expected {
			t.Errorf("expected %q, actual %q", expected, actual)
		}
	}
}

func TestFixCodefence(t *testing.T) {
	table := [][]string{
		{"```{\n  x: y,\n  a: b\n}```", "```\n{\n  x: y,\n  a: b\n}\n```"},
	}

	for _, test := range table {
		input := test[0]
		expected := test[1]
		actual := openCodefence.ReplaceAllLiteralString(input, "```\n")
		actual = closeCodefence.ReplaceAllString(actual, "$1\n```")

		if actual != expected {
			t.Errorf("expected %q, actual %q", expected, actual)
		}
	}
}
