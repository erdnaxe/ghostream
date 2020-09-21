/*
 * Copyright (C) 2020 Cr@ns <roots@crans.org>
 * Authors : Alexandre Iooss <erdnaxe@crans.org>
 * SPDX-License-Identifier: MIT
 */

package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

// Config holds app configuration
type Config struct {
	AuthBackend string
	LDAP        struct {
		URI    string
		UserDn string
	}
	Prometheus struct {
		ListenAddress string
	}
	Site struct {
		ListenAddress string
		Name          string
		Hostname      string
		Favicon       string
		WidgetURL     string
	}
}

// New configuration
func New() (*Config, error) {
	// Load configuration from environnement variables
	// Replace "." to "_" for nested structs
	// e.g. GHOSTREAM_LDAP_URI will apply to Config.LDAP.URI
	viper.SetEnvPrefix("ghostream")
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv()

	// Load configuration file if exists
	viper.SetConfigName("ghostream")
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
			log.Fatal(err)
		}
	} else {
		// Config loaded
		log.Printf("Using config file: %s", viper.ConfigFileUsed())
	}

	// Define configuration default values
	viper.SetDefault("AuthBackend", "LDAP")
	viper.SetDefault("LDAP.URI", "ldap://127.0.0.1:389")
	viper.SetDefault("LDAP.UserDn", "cn=users,dc=example,dc=com")
	viper.SetDefault("Prometheus.ListenAddress", "0.0.0.0:2112")
	viper.SetDefault("Site.ListenAddress", "127.0.0.1:8080")
	viper.SetDefault("Site.Name", "Ghostream")
	viper.SetDefault("Site.Hostname", "localhost")
	viper.SetDefault("Site.Favicon", "/favicon.ico")

	config := &Config{}
	err := viper.Unmarshal(config)
	return config, err
}
