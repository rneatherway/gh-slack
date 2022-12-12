package main

import "testing"

func TestPermalinkRE(t *testing.T) {
	result := permalinkRE.FindStringSubmatch("https://github.slack.com/archives/CP9GMKJCE/p1648028606962719")
	if len(result) != 5 {
		t.Errorf("result had length %d", len(result))
	}
	if result[1] != "github" || result[2] != "CP9GMKJCE" || result[3] != "1648028606" || result[4] != "962719" {
		t.Fail()
	}
}
