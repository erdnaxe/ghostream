package main

import (
	"github.com/spf13/viper"
	"gitlab.crans.org/nounous/ghostream/auth"
	"gitlab.crans.org/nounous/ghostream/internal/monitoring"
	"gitlab.crans.org/nounous/ghostream/stream/multicast"
	"gitlab.crans.org/nounous/ghostream/stream/srt"
	"gitlab.crans.org/nounous/ghostream/stream/webrtc"
	"gitlab.crans.org/nounous/ghostream/web"
	"testing"
)

// TestLoadConfiguration tests the configuration file loading and the init of some parameters
func TestLoadConfiguration(t *testing.T) {
	loadConfiguration()
	cfg := struct {
		Auth       auth.Options
		Monitoring monitoring.Options
		Multicast  multicast.Options
		Srt        srt.Options
		Web        web.Options
		WebRTC     webrtc.Options
	}{}
	if err := viper.Unmarshal(&cfg); err != nil {
		t.Fatal("Failed to load settings", err)
	}
}
