// Package transcoder manages transcoders
package transcoder

import (
	"gitlab.crans.org/nounous/ghostream/messaging"
	"gitlab.crans.org/nounous/ghostream/transcoder/text"
)

// Options holds text package configuration
type Options struct {
	Text text.Options
}

// Init all transcoders
func Init(streams *messaging.Streams, cfg *Options) {
	go text.Init(streams, &cfg.Text)
}
