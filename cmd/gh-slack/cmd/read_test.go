package cmd

import "testing"

func TestParsePermalink(t *testing.T) {
	tests := []struct {
		link     string
		expected linkParts
	}{
		{
			link: "https://github.slack.com/archives/CP9GMKJCE/p1648028606962719",
			expected: linkParts{
				team:      "github",
				channelID: "CP9GMKJCE",
				timestamp: "1648028606.962719",
			},
		},
		{
			link: "https://sanity-io-land.slack.com/archives/C9Y51FDGA/p1709663536325529",
			expected: linkParts{
				team:      "sanity-io-land",
				channelID: "C9Y51FDGA",
				timestamp: "1709663536.325529",
			},
		},
		{
			link: "https://example.slack.com/archives/ABC123/p1709663536325529?thread_ts=1234567890.123456&cid=ABC123",
			expected: linkParts{
				team:      "example",
				channelID: "ABC123",
				thread:    "1234567890.123456",
				timestamp: "1709663536.325529",
			},
		},
	}

	for _, test := range tests {
		actual, err := parsePermalink(test.link)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if actual.team != test.expected.team || actual.channelID != test.expected.channelID || actual.timestamp != test.expected.timestamp {
			t.Errorf("unexpected result for link %s, got %+v, want %+v", test.link, actual, test.expected)
		}
	}
}
