package srt

import (
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

	for {
		s, err := sck.Accept()
		if err != nil {
			log.Println("Error occured while accepting request:", err)
			continue
		}

		go func(s *sck.SrtSocket) {
			buff := make([]byte, 2048)
			for {
				n, err := s.Read(buff, 10000)
				if err != nil {
					log.Println("Error occured while reading SRT socket:", err)
					break
				}
				if n == 0 {
					// End of stream
					break
				}
				log.Printf("Received %d bytes", n)
			}
		}(s)
	}
}
