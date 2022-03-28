package main

import "testing"

func TestRE(t *testing.T) {
	result := permalinkRE.FindStringSubmatch("https://github.slack.com/archives/CP9GMKJCE/p1648028606962719")
	for _, r := range result {
		t.Log(r)
	}
	t.Fail()
}
