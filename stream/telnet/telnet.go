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
	// Cfg contains the different options of the telnet package, see below
	// TODO Config should not be exported
	Cfg            *Options
	currentMessage map[string]*string
	clientCount    map[string]int

	asciiChars = []byte{' ', '.', ',', ':', ';', '+', '*', '?', '%', '$', '#', '@'}
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

	if !config.Enabled {
		return
	}

	currentMessage = make(map[string]*string)
	clientCount = make(map[string]int)

	listener, err := net.Listen("tcp", config.ListenAddress)
	if err != nil {
		log.Printf("Error while listening to the address %s: %s", config.ListenAddress, err)
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
					_, err = s.Write([]byte("[GHOSTREAM]\nEnter stream ID: "))
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

					// Avoid bruteforce
					time.Sleep(3 * time.Second)

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

				clientCount[streamID]++

				for {
					n, err := s.Write([]byte(*currentMessage[streamID]))
					if err != nil {
						log.Printf("Error while sending TCP data: %s", err)
						_ = s.Close()
						clientCount[streamID]--
						break
					}
					if n == 0 {
						_ = s.Close()
						clientCount[streamID]--
						break
					}
					time.Sleep(time.Duration(config.Delay) * time.Millisecond)
				}
			}(s)
		}
	}()

	log.Println("Telnet server initialized")
}

// GetNumberConnectedSessions returns the numbers of clients that are viewing the stream through a telnet shell
func GetNumberConnectedSessions(streamID string) int {
	if Cfg == nil || !Cfg.Enabled {
		return 0
	}
	return clientCount[streamID]
}

// StartASCIIArtStream send all packets received by ffmpeg as ASCII Art to telnet clients
func StartASCIIArtStream(streamID string, reader io.ReadCloser) {
	if !Cfg.Enabled {
		_ = reader.Close()
		return
	}

	currentMessage[streamID] = new(string)
	pixelBuff := make([]byte, Cfg.Width*Cfg.Height)
	textBuff := strings.Builder{}
	for {
		n, err := reader.Read(pixelBuff)
		if err != nil {
			log.Printf("An error occured while reading input: %s", err)
			break
		}
		if n == 0 {
			// Stream is finished
			break
		}

		// Header
		textBuff.Reset()
		textBuff.Grow((2*Cfg.Width + 1) * Cfg.Height)
		for i := 0; i < 42; i++ {
			textBuff.WriteByte('\n')
		}

		// Convert image to ASCII
		for j := 0; j < Cfg.Height; j++ {
			for i := 0; i < Cfg.Width; i++ {
				textBuff.WriteByte(asciiChars[pixelBuff[Cfg.Width*j+i]/22])
				textBuff.WriteByte(asciiChars[pixelBuff[Cfg.Width*j+i]/22])
			}
			textBuff.WriteByte('\n')
		}

		*(currentMessage[streamID]) = textBuff.String()
	}
}
