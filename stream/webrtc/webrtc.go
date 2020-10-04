package webrtc

import (
	"bufio"
	"fmt"
	"github.com/pion/rtp"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/ivfreader"
	"github.com/pion/webrtc/v3/pkg/media/oggreader"
	"gitlab.crans.org/nounous/ghostream/internal/monitoring"
	"gitlab.crans.org/nounous/ghostream/stream/srt"
)

// Options holds web package configuration
type Options struct {
	MinPortUDP uint16
	MaxPortUDP uint16

	STUNServers []string
}

// SessionDescription contains SDP data
// to initiate a WebRTC connection between one client and this app
type SessionDescription = webrtc.SessionDescription

const (
	audioFileName = "output.ogg"
	videoFileName = "toto.ivf"
)

var (
	videoTracks []*webrtc.Track
	audioTracks []*webrtc.Track
)

// Helper to reslice tracks
func removeTrack(tracks []*webrtc.Track, track *webrtc.Track) []*webrtc.Track {
	for i, t := range tracks {
		if t == track {
			return append(tracks[:i], tracks[i+1:]...)
		}
	}
	return nil
}

// GetNumberConnectedSessions get the number of currently connected clients
func GetNumberConnectedSessions() int {
	return len(videoTracks)
}

// newPeerHandler is called when server receive a new session description
// this initiates a WebRTC connection and return server description
func newPeerHandler(remoteSdp webrtc.SessionDescription, cfg *Options) webrtc.SessionDescription {
	// Create media engine using client SDP
	mediaEngine := webrtc.MediaEngine{}
	if err := mediaEngine.PopulateFromSDP(remoteSdp); err != nil {
		log.Println("Failed to create new media engine", err)
		return webrtc.SessionDescription{}
	}

	// Create a new PeerConnection
	settingsEngine := webrtc.SettingEngine{}
	if err := settingsEngine.SetEphemeralUDPPortRange(cfg.MinPortUDP, cfg.MaxPortUDP); err != nil {
		log.Println("Failed to set min/max UDP ports", err)
		return webrtc.SessionDescription{}
	}
	api := webrtc.NewAPI(
		webrtc.WithMediaEngine(mediaEngine),
		webrtc.WithSettingEngine(settingsEngine),
	)
	peerConnection, err := api.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{{URLs: cfg.STUNServers}},
	})
	if err != nil {
		log.Println("Failed to initiate peer connection", err)
		return webrtc.SessionDescription{}
	}

	// Create video track
	codec, payloadType := getPayloadType(mediaEngine, webrtc.RTPCodecTypeVideo, "VP8")
	videoTrack, err := webrtc.NewTrack(payloadType, rand.Uint32(), "video", "pion", codec)
	if err != nil {
		log.Println("Failed to create new video track", err)
		return webrtc.SessionDescription{}
	}
	if _, err = peerConnection.AddTrack(videoTrack); err != nil {
		log.Println("Failed to add video track", err)
		return webrtc.SessionDescription{}
	}

	// Create audio track
	codec, payloadType = getPayloadType(mediaEngine, webrtc.RTPCodecTypeAudio, "opus")
	audioTrack, err := webrtc.NewTrack(payloadType, rand.Uint32(), "audio", "pion", codec)
	if err != nil {
		log.Println("Failed to create new audio track", err)
		return webrtc.SessionDescription{}
	}
	if _, err = peerConnection.AddTrack(audioTrack); err != nil {
		log.Println("Failed to add audio track", err)
		return webrtc.SessionDescription{}
	}

	// Set the remote SessionDescription
	if err = peerConnection.SetRemoteDescription(remoteSdp); err != nil {
		log.Println("Failed to set remote description", err)
		return webrtc.SessionDescription{}
	}

	// Create answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		log.Println("Failed to create answer", err)
		return webrtc.SessionDescription{}
	}

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	// Sets the LocalDescription, and starts our UDP listeners
	if err = peerConnection.SetLocalDescription(answer); err != nil {
		log.Println("Failed to set local description", err)
		return webrtc.SessionDescription{}
	}

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		log.Printf("Connection State has changed %s \n", connectionState.String())
		if connectionState == webrtc.ICEConnectionStateConnected {
			// Register tracks
			videoTracks = append(videoTracks, videoTrack)
			audioTracks = append(audioTracks, audioTrack)
			monitoring.WebRTCConnectedSessions.Inc()
		} else if connectionState == webrtc.ICEConnectionStateDisconnected {
			// Unregister tracks
			videoTracks = removeTrack(videoTracks, videoTrack)
			audioTracks = removeTrack(audioTracks, audioTrack)
			monitoring.WebRTCConnectedSessions.Dec()
		}
	})

	// Block until ICE Gathering is complete, disabling trickle ICE
	// we do this because we only can exchange one signaling message
	// in a production application you should exchange ICE Candidates via OnICECandidate
	<-gatherComplete

	// Output the local description and send it to browser
	return *peerConnection.LocalDescription()
}

func playVideo() {
	// Open a IVF file and start reading using our IVFReader
	file, ivfErr := os.Open(videoFileName)
	if ivfErr != nil {
		panic(ivfErr)
	}

	ivf, header, ivfErr := ivfreader.NewWith(file)
	if ivfErr != nil {
		panic(ivfErr)
	}

	// Send our video file frame at a time. Pace our sending so we send it at the same speed it should be played back as.
	// This isn't required since the video is timestamped, but we will such much higher loss if we send all at once.
	sleepTime := time.Millisecond * time.Duration((float32(header.TimebaseNumerator)/float32(header.TimebaseDenominator))*1000)
	for {
		// Need at least one client
		frame, _, ivfErr := ivf.ParseNextFrame()
		if ivfErr == io.EOF {
			fmt.Printf("All video frames parsed and sent")
			os.Exit(0)
		}

		if ivfErr != nil {
			panic(ivfErr)
		}

		time.Sleep(sleepTime)
		for _, t := range videoTracks {
			if ivfErr = t.WriteSample(media.Sample{Data: frame, Samples: 90000}); ivfErr != nil {
				log.Fatalln("Failed to write video stream:", ivfErr)
			}
		}
	}
}

func playAudio() {
	// Open a IVF file and start reading using our IVFReader
	file, oggErr := os.Open(audioFileName)
	if oggErr != nil {
		panic(oggErr)
	}

	// Open on oggfile in non-checksum mode.
	ogg, _, oggErr := oggreader.NewWith(file)
	if oggErr != nil {
		panic(oggErr)
	}

	// Keep track of last granule, the difference is the amount of samples in the buffer
	var lastGranule uint64
	for {
		// Need at least one client
		pageData, pageHeader, oggErr := ogg.ParseNextPage()
		if oggErr == io.EOF {
			fmt.Printf("All audio pages parsed and sent")
			os.Exit(0)
		}

		if oggErr != nil {
			panic(oggErr)
		}

		// The amount of samples is the difference between the last and current timestamp
		sampleCount := float64(pageHeader.GranulePosition - lastGranule)
		lastGranule = pageHeader.GranulePosition

		for _, t := range audioTracks {
			if oggErr = t.WriteSample(media.Sample{Data: pageData, Samples: uint32(sampleCount)}); oggErr != nil {
				log.Fatalln("Failed to write audio stream:", oggErr)
			}
		}

		// Convert seconds to Milliseconds, Sleep doesn't accept floats
		time.Sleep(time.Duration((sampleCount/48000)*1000) * time.Millisecond)
	}
}

// Search for Codec PayloadType
//
// Since we are answering we need to match the remote PayloadType
func getPayloadType(m webrtc.MediaEngine, codecType webrtc.RTPCodecType, codecName string) (*webrtc.RTPCodec, uint8) {
	for _, codec := range m.GetCodecsByKind(codecType) {
		if codec.Name == codecName {
			return codec, codec.PayloadType
		}
	}
	panic(fmt.Sprintf("Remote peer does not support %s", codecName))
}

func waitForPackets(inputChannel chan srt.Packet) {
	// FIXME Clean code
	var ffmpeg *exec.Cmd
	var ffmpegInput io.WriteCloser
	for {
		var err error = nil
		packet := <-inputChannel
		switch packet.PacketType {
		case "register":
			log.Printf("WebRTC RegisterStream %s", packet.StreamName)

			// Copied from https://github.com/pion/webrtc/blob/master/examples/rtp-to-webrtc/main.go

			// Open a UDP Listener for RTP Packets on port 5004
			videoListener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 5004})
			if err != nil {
				panic(err)
			}
			audioListener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 5005})
			if err != nil {
				panic(err)
			}
			defer func() {
				if err = videoListener.Close(); err != nil {
					panic(err)
				}
				if err = audioListener.Close(); err != nil {
					panic(err)
				}
			}()

			ffmpeg = exec.Command("ffmpeg", "-re", "-i", "pipe:0",
				"-an", "-vcodec", "libvpx", //"-cpu-used", "5", "-deadline", "1", "-g", "10", "-error-resilient", "1", "-auto-alt-ref", "1",
				"-f", "rtp", "rtp://127.0.0.1:5004",
				"-vn", "-acodec", "libopus", //"-cpu-used", "5", "-deadline", "1", "-g", "10", "-error-resilient", "1", "-auto-alt-ref", "1",
				"-f", "rtp", "rtp://127.0.0.1:5005")

			fmt.Println("Waiting for RTP Packets, please run GStreamer or ffmpeg now")

			input, err := ffmpeg.StdinPipe()
			if err != nil {
				panic(err)
			}
			ffmpegInput = input
			errOutput, err := ffmpeg.StderrPipe()
			if err != nil {
				panic(err)
			}

			if err := ffmpeg.Start(); err != nil {
				panic(err)
			}

			// Receive video
			go func() {
				for {
					// Listen for a single RTP Packet, we need this to determine the SSRC
					inboundRTPPacket := make([]byte, 1500) // UDP MTU
					n, _, err := videoListener.ReadFromUDP(inboundRTPPacket)
					if err != nil {
						panic(err)
					}
					packet := &rtp.Packet{}
					if err := packet.Unmarshal(inboundRTPPacket[:n]); err != nil {
						panic(err)
					}
					log.Printf("[Video] %s", packet)
					for _, videoTrack := range videoTracks {
						if writeErr := videoTrack.WriteRTP(packet); writeErr != nil {
							panic(err)
						}
					}
				}
			}()

			// Receive audio
			go func() {
				for {
					// Listen for a single RTP Packet, we need this to determine the SSRC
					inboundRTPPacket := make([]byte, 1500) // UDP MTU
					n, _, err := audioListener.ReadFromUDP(inboundRTPPacket)
					if err != nil {
						panic(err)
					}
					packet := &rtp.Packet{}
					if err := packet.Unmarshal(inboundRTPPacket[:n]); err != nil {
						panic(err)
					}
					log.Printf("[Audio] %s", packet)
					for _, audioTrack := range audioTracks {
						if writeErr := audioTrack.WriteRTP(packet); writeErr != nil {
							panic(err)
						}
					}
				}
			}()

			go func() {
				scanner := bufio.NewScanner(errOutput)
				for scanner.Scan() {
					log.Printf("[WEBRTC FFMPEG %s] %s", "demo", scanner.Text())
				}
			}()
			break
		case "sendData":
			// log.Printf("WebRTC SendPacket %s", packet.StreamName)
			_, err := ffmpegInput.Write(packet.Data)
			if err != nil {
				panic(err)
			}
			break
		case "close":
			log.Printf("WebRTC CloseConnection %s", packet.StreamName)
			break
		default:
			log.Println("Unknown SRT packet type:", packet.PacketType)
			break
		}
		if err != nil {
			log.Printf("Error occured while receiving SRT packet of type %s: %s", packet.PacketType, err)
		}
	}
}

// Serve WebRTC media streaming server
func Serve(remoteSdpChan, localSdpChan chan webrtc.SessionDescription, inputChannel chan srt.Packet, cfg *Options) {
	log.Printf("WebRTC server using UDP from %d to %d", cfg.MinPortUDP, cfg.MaxPortUDP)

	// FIXME: use data from inputChannel
	go waitForPackets(inputChannel)
	// go playVideo()
	// go playAudio()

	// Handle new connections
	for {
		// Wait for incoming session description
		// then send the local description to browser
		localSdpChan <- newPeerHandler(<-remoteSdpChan, cfg)
	}
}
