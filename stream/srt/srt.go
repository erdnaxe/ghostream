package srt

import (
	"log"
	"net"

	"github.com/openfresh/gosrt/srt"
)

// Options holds web package configuration
type Options struct {
	ListenAddress string
}

// Serve SRT server
func Serve(cfg *Options) {
	log.Printf("SRT server listening on %s", cfg.ListenAddress)
	l, _ := srt.Listen("srt", cfg.ListenAddress)
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("Error on incoming SRT stream", err)
		}
		log.Printf("New incomming SRT stream from %s", conn.RemoteAddr())

		go func(sc net.Conn) {
			defer sc.Close()

			for {
				//mon := conn.(*srt.SRTConn).Stats()
				//s, _ := json.MarshalIndent(mon, "", "\t")
				//fmt.Println(string(s))
			}
		}(conn)
	}
}
