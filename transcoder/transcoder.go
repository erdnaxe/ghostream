package transcoder

import (
	"gitlab.crans.org/nounous/ghostream/stream"
	"gitlab.crans.org/nounous/ghostream/transcoder/text"
)

// Options holds text package configuration
type Options struct {
	Text text.Options
}

// Init all transcoders
func Init(streams map[string]*stream.Stream, cfg *Options) {
	go text.Init(streams, &cfg.Text)
}
