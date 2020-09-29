package srt

import (
	"testing"
)

func TestSplitHostPort(t *testing.T) {
	host, port := splitHostPort("127.0.0.1:1234")
	if host != "127.0.0.1" || port != 1234 {
		t.Errorf("splitHostPort returned %v:%d != 127.0.0.1:1234", host, port)
	}

	host, port = splitHostPort(":1234")
	if host != "0.0.0.0" || port != 1234 {
		t.Errorf("splitHostPort returned %v:%d != 0.0.0.0:1234", host, port)
	}
}
