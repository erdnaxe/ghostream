// Package config loads application settings
package config

import (
	"bytes"
	"log"
	"net"
	"strings"

	"github.com/spf13/viper"
	"gitlab.crans.org/nounous/ghostream/auth"
	"gitlab.crans.org/nounous/ghostream/auth/basic"
	"gitlab.crans.org/nounous/ghostream/auth/ldap"
	"gitlab.crans.org/nounous/ghostream/internal/monitoring"
	"gitlab.crans.org/nounous/ghostream/stream/forwarding"
	"gitlab.crans.org/nounous/ghostream/stream/srt"
	"gitlab.crans.org/nounous/ghostream/stream/webrtc"
	"gitlab.crans.org/nounous/ghostream/web"
	"gopkg.in/yaml.v2"
)

// Config holds application configuration
type Config struct {
	Auth       auth.Options
	Forwarding forwarding.Options
	Monitoring monitoring.Options
	Srt        srt.Options
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
				Credentials: map[string]string{
					// Demo user with password "demo"
					"demo": "$2b$15$LRnG3eIHFlYIguTxZOLH7eHwbQC/vqjnLq6nDFiHSUDKIU.f5/1H6",
				},
			},
			LDAP: ldap.Options{
				URI:    "ldap://127.0.0.1:389",
				UserDn: "cn=users,dc=example,dc=com",
			},
		},
		Forwarding: make(map[string][]string),
		Monitoring: monitoring.Options{
			Enabled:       true,
			ListenAddress: ":2112",
		},
		Srt: srt.Options{
			Enabled:       true,
			ListenAddress: ":9710",
			MaxClients:    64,
		},
		Web: web.Options{
			Enabled:                     true,
			Favicon:                     "/static/img/favicon.svg",
			Hostname:                    "localhost",
			ListenAddress:               ":8080",
			Name:                        "Ghostream",
			OneStreamPerDomain:          false,
			PlayerPoster:                "/static/img/no_stream.svg",
			ViewersCounterRefreshPeriod: 20000,
		},
		WebRTC: webrtc.Options{
			Enabled:     true,
			MaxPortUDP:  10005,
			MinPortUDP:  10000,
			STUNServers: []string{"stun:stun.l.google.com:19302"},
		},
	}
}

// Load global configuration as a struct
func Load() (*Config, error) {
	// Viper needs to know if a key exists in order to override it.
	// See https://github.com/spf13/viper/issues/188
	b, err := yaml.Marshal(New())
	if err != nil {
		return nil, err
	}
	defaultConfig := bytes.NewReader(b)
	viper.SetConfigType("yaml")
	if err := viper.MergeConfig(defaultConfig); err != nil {
		return nil, err
	}

	// Overwrite configuration from file if exists
	viper.SetConfigName("ghostream.yml")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.ghostream")
	viper.AddConfigPath("/etc/ghostream")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, ignore and use defaults
			log.Print(err)
		} else {
			// Config file was found but another error was produced
			return nil, err
		}
	} else {
		// Config loaded
		log.Printf("Using config file: %s", viper.ConfigFileUsed())
	}

	// Overwrite configuration from environnement variables
	// Replace "." to "_" for nested structs
	// e.g. GHOSTREAM_LDAP_URI will apply to Config.LDAP.URI
	viper.AutomaticEnv()
	viper.SetEnvPrefix("ghostream")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Load onto struct
	cfg := &Config{}
	if err := viper.UnmarshalExact(cfg); err != nil {
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

	return cfg, nil
}
