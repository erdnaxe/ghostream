package main

import (
	"log"
	"strings"

	"github.com/spf13/viper"
	"gitlab.crans.org/nounous/ghostream/auth"
	"gitlab.crans.org/nounous/ghostream/internal/monitoring"
	"gitlab.crans.org/nounous/ghostream/web"
)

func loadConfiguration() {
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
	viper.SetDefault("Auth.Backend", "LDAP")
	viper.SetDefault("Auth.LDAP.URI", "ldap://127.0.0.1:389")
	viper.SetDefault("Auth.LDAP.UserDn", "cn=users,dc=example,dc=com")
	viper.SetDefault("Monitoring.ListenAddress", ":2112")
	viper.SetDefault("Web.ListenAddress", "127.0.0.1:8080")
	viper.SetDefault("Web.Name", "Ghostream")
	viper.SetDefault("Web.Hostname", "localhost")
	viper.SetDefault("Web.Favicon", "/favicon.ico")
}

func main() {
	// Load configuration
	loadConfiguration()
	cfg := struct {
		Auth       auth.Options
		Monitoring monitoring.Options
		Web        web.Options
	}{}
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalln("Failed to load settings", err)
	}

	// Init authentification
	authBackend, err := auth.New(&cfg.Auth)
	if err != nil {
		log.Fatalln("Failed to load authentification backend:", err)
	}
	defer authBackend.Close()

	// Start web server routine
	go func() {
		web.ServeHTTP(&cfg.Web)
	}()

	// Start monitoring server routine
	go func() {
		monitoring.ServeHTTP(&cfg.Monitoring)
	}()

	// Wait for routines
	select {}
}