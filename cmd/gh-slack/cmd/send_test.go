package cmd

import "testing"

func TestTargetFromThread(t *testing.T) {
	tests := []struct {
		thread        string
		expected      messageTarget
		fromPermalink bool
	}{
		{
			thread: "1712345678.123456",
			expected: messageTarget{
				threadTS: "1712345678.123456",
			},
		},
		{
			thread: "https://example.slack.com/archives/C123/p1712345678123456",
			expected: messageTarget{
				team:      "example",
				channelID: "C123",
				threadTS:  "1712345678.123456",
			},
			fromPermalink: true,
		},
		{
			thread: "https://example.slack.com/archives/C123/p1712345680123456" +
				"?thread_ts=1712345678.123456&cid=C123",
			expected: messageTarget{
				team:      "example",
				channelID: "C123",
				threadTS:  "1712345678.123456",
			},
			fromPermalink: true,
		},
	}

	for _, test := range tests {
		actual, fromPermalink, err := targetFromThread(test.thread)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", test.thread, err)
		}
		if actual != test.expected || fromPermalink != test.fromPermalink {
			t.Fatalf("unexpected target for %q: %+v, %t", test.thread, actual, fromPermalink)
		}
	}
}
