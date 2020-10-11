package config

import (
	"testing"
)

func TestLoad(t *testing.T) {
	_, err := Load()
	if err != nil {
		t.Error("Failed to load configuration:", err)
	}
}
