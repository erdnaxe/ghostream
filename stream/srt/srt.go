package srt

import (
	"fmt"
	"log"

	"github.com/haivision/srtgo"
)

// Options holds web package configuration
type Options struct {
	ListenAddress string
}

// Serve SRT server
func Serve(cfg *Options) {
	log.Printf("SRT server listening on %s", cfg.ListenAddress)

	options := make(map[string]string)
	options["transtype"] = "file"

	// FIXME: cfg.ListenAddress -> host and port
	sck := srtgo.NewSrtSocket("0.0.0.0", 9710, options)
	sck.Listen(1)
	s, _ := sck.Accept()

	buff := make([]byte, 2048)
	for {
		n, _ := s.Read(buff, 10000)
		if n == 0 {
			break
		}
		fmt.Printf("Received %d bytes", n)
	}
}
