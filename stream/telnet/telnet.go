// Package telnet provides some fancy tools, like an ASCII-art stream.
package telnet

import (
	"io"
	"log"
	"net"
	"time"
)

var (
	Cfg            *Options
	currentMessage *string
)

// Options holds telnet package configuration
type Options struct {
	Enabled       bool
	ListenAddress string
	Width         int
	Height        int
	Delay         int
}

func Serve(config *Options) {
	Cfg = config

	if !Cfg.Enabled {
		return
	}

	listener, err := net.Listen("tcp", Cfg.ListenAddress)
	if err != nil {
		log.Printf("Error while listening to the address %s: %s", Cfg.ListenAddress, err)
		return
	}

	currentMessage = new(string)

	go func() {
		for {
			s, err := listener.Accept()
			if err != nil {
				log.Printf("Error while accepting TCP socket: %s", s)
				continue
			}
			go func(s net.Conn) {
				for {
					n, err := s.Write([]byte(*currentMessage))
					if err != nil {
						log.Printf("Error while sending TCP data: %s", err)
						_ = s.Close()
						break
					}
					if n == 0 {
						_ = s.Close()
						break
					}
					time.Sleep(time.Duration(Cfg.Delay) * time.Millisecond)
				}
			}(s)
		}
	}()

	log.Println("Telnet server initialized")
}

func asciiChar(pixel byte) string {
	asciiChars := []string{"@", "#", "$", "%", "?", "*", "+", ";", ":", ",", ".", " "}
	return asciiChars[(255-pixel)/23]
}

// ServeAsciiArt starts a telnet server that send all packets as ASCII Art
func ServeAsciiArt(reader io.ReadCloser) {
	if !Cfg.Enabled {
		_ = reader.Close()
		return
	}

	buff := make([]byte, Cfg.Width*Cfg.Height)
	header := "\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n"
	for {
		n, _ := reader.Read(buff)
		if n == 0 {
			break
		}
		imageStr := ""
		for j := 0; j < Cfg.Height; j++ {
			for i := 0; i < Cfg.Width; i++ {
				pixel := buff[Cfg.Width*j+i]
				imageStr += asciiChar(pixel) + asciiChar(pixel)
			}
			imageStr += "\n"
		}
		*currentMessage = header + imageStr
	}
}
