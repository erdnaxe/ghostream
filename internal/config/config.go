// Package config loads application settings
package config

import (
	"gitlab.crans.org/nounous/ghostream/stream/ovenmediaengine"
	"net"

	"github.com/sherifabdlnaby/configuro"
	"gitlab.crans.org/nounous/ghostream/auth"
	"gitlab.crans.org/nounous/ghostream/auth/basic"
	"gitlab.crans.org/nounous/ghostream/auth/ldap"
	"gitlab.crans.org/nounous/ghostream/internal/monitoring"
	"gitlab.crans.org/nounous/ghostream/stream/forwarding"
	"gitlab.crans.org/nounous/ghostream/stream/srt"
	"gitlab.crans.org/nounous/ghostream/stream/telnet"
	"gitlab.crans.org/nounous/ghostream/stream/webrtc"
	"gitlab.crans.org/nounous/ghostream/transcoder"
	"gitlab.crans.org/nounous/ghostream/transcoder/text"
	"gitlab.crans.org/nounous/ghostream/web"
)

// Config holds application configuration
type Config struct {
	Auth       auth.Options
	Forwarding forwarding.Options
	Monitoring monitoring.Options
	OME        ovenmediaengine.Options
	Srt        srt.Options
	Telnet     telnet.Options
	Transcoder transcoder.Options
	Web        web.Options
	WebRTC     webrtc.Options
}

// New configuration with default values
func New() *Config {
	return &Config{
		Auth: auth.Options{
			Enabled: true,
			Backend: "Basic",
			Basic: basic.Options{
				Credentials: make(map[string]string),
			},
			LDAP: ldap.Options{
				Aliases: make(map[string]string),
				URI:     "ldap://127.0.0.1:389",
				UserDn:  "cn=users,dc=example,dc=com",
			},
		},
		Forwarding: make(map[string][]string),
		Monitoring: monitoring.Options{
			Enabled:       true,
			ListenAddress: ":2112",
		},
		OME: ovenmediaengine.Options{
			Enabled: true,
			URL:     "ovenmediaengine:1915",
			App:     "play",
		},
		Srt: srt.Options{
			Enabled:       true,
			ListenAddress: ":9710",
			MaxClients:    64,
		},
		Telnet: telnet.Options{
			Enabled:       false,
			ListenAddress: ":8023",
		},
		Transcoder: transcoder.Options{
			Text: text.Options{
				Enabled:   false,
				Width:     80,
				Height:    45,
				Framerate: 20,
			},
		},
		Web: web.Options{
			Enabled:                     true,
			Favicon:                     "/static/img/favicon.svg",
			Hostname:                    "localhost",
			ListenAddress:               ":8080",
			Name:                        "Ghostream",
			MapDomainToStream:           make(map[string]string),
			PlayerPoster:                "/static/img/no_stream.svg",
			ViewersCounterRefreshPeriod: 20000,
			LegalMentionsEntity:         "l'association Crans",
			LegalMentionsAddress:        "61 Avenue du Président Wilson, 94235 Cachan Cedex, France",
			LegalMentionsFullAddress: []string{"Association Cr@ns - ENS Paris-Saclay",
				"Notification de Contenus Illicites", "4, avenue des Sciences", "91190 Gif-sur-Yvette", "France"},
			LegalMentionsEmail: "bureau[at]crans.org",
		},
		WebRTC: webrtc.Options{
			Enabled:     false,
			MaxPortUDP:  11000,
			MinPortUDP:  10000,
			STUNServers: []string{"stun:stun.l.google.com:19302"},
		},
	}
}

// Load global configuration as a struct
func Load() (*Config, error) {
	// Create Configuro
	config, err := configuro.NewConfig(
		configuro.WithLoadFromEnvVars("GHOSTREAM"),
		configuro.WithLoadFromConfigFile("/etc/ghostream/ghostream.yml", false),
		configuro.WithEnvConfigPathOverload("GHOSTREAM_CONFIG"),
	)
	if err != nil {
		return nil, err
	}

	// Load default configuration
	cfg := New()

	// Load values in configuration struct
	if err := config.Load(cfg); err != nil {
		return nil, err
	}

	// Copy STUN configuration to clients
	cfg.Web.STUNServers = cfg.WebRTC.STUNServers

	// Copy SRT server port to display it on web page
	_, srtPort, err := net.SplitHostPort(cfg.Srt.ListenAddress)
	if err != nil {
		return nil, err
	}
	cfg.Web.SRTServerPort = srtPort

	// If no credentials register, add demo account with password "demo"
	if len(cfg.Auth.Basic.Credentials) < 1 {
		cfg.Auth.Basic.Credentials["demo"] = "$2b$10$xuU7XFwmRX2CMgdSaA8rM.4Y8.BtRNzhUedwN0G8tCegDRNUERTCS"
	}

	return cfg, nil
}
