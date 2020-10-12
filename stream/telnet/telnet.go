// Package telnet provides some fancy tools, like an ASCII-art stream.
package telnet

import (
	"io"
	"log"
	"net"
	"strings"
	"time"
)

var (
	// TODO Config should not be exported
	// Cfg contains the different options of the telnet package, see below
	Cfg            *Options
	currentMessage map[string]string
)

// Options holds telnet package configuration
type Options struct {
	Enabled       bool
	ListenAddress string
	Width         int
	Height        int
	Delay         int
}

// Serve starts the telnet server and listen to clients
func Serve(config *Options) {
	Cfg = config

	if !Cfg.Enabled {
		return
	}

	currentMessage = make(map[string]string)

	listener, err := net.Listen("tcp", Cfg.ListenAddress)
	if err != nil {
		log.Printf("Error while listening to the address %s: %s", Cfg.ListenAddress, err)
		return
	}

	go func() {
		for {
			s, err := listener.Accept()
			if err != nil {
				log.Printf("Error while accepting TCP socket: %s", s)
				continue
			}

			go func(s net.Conn) {
				streamID := ""
				// Request for stream ID
				for {
					_, _ = s.Write([]byte("[GHOSTREAM]\n"))
					_, err = s.Write([]byte("Enter stream ID: "))
					if err != nil {
						log.Println("Error while requesting stream ID to telnet client")
						_ = s.Close()
						return
					}
					buff := make([]byte, 255)
					n, err := s.Read(buff)
					if err != nil {
						log.Println("Error while requesting stream ID to telnet client")
						_ = s.Close()
						return
					}

					streamID = string(buff[:n])
					streamID = strings.Replace(streamID, "\r", "", -1)
					streamID = strings.Replace(streamID, "\n", "", -1)

					if len(streamID) > 0 {
						if strings.ToLower(streamID) == "exit" {
							_, _ = s.Write([]byte("Goodbye!\n"))
							_ = s.Close()
							return
						}
						if _, ok := currentMessage[streamID]; !ok {
							_, err = s.Write([]byte("Unknown stream ID.\n"))
							if err != nil {
								log.Println("Error while requesting stream ID to telnet client")
								_ = s.Close()
								return
							}
							continue
						}
						break
					}
				}

				for {
					n, err := s.Write([]byte(currentMessage[streamID]))
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
	return asciiChars[(255-pixel)/22]
}

// StartASCIIArtStream send all packets received by ffmpeg as ASCII Art to telnet clients
func StartASCIIArtStream(streamID string, reader io.ReadCloser) {
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
		currentMessage[streamID] = header + imageStr
	}
}
