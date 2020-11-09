// Package ovenmediaengine provides the forwarding to an ovenmediaengine server to handle the web client
package ovenmediaengine

import "gitlab.crans.org/nounous/ghostream/messaging"

// Options holds ovenmediaengine package configuration
type Options struct {
	Enabled bool
	URL     string
	App     string
}

func Serve(streams *messaging.Streams, cfg *Options) {
	if !cfg.Enabled {
		return
	}
}
