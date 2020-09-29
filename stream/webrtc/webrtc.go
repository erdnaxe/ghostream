package webrtc

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/ivfreader"
	"github.com/pion/webrtc/v3/pkg/media/oggreader"
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
	videoFileName = "output.ivf"
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
		} else if connectionState == webrtc.ICEConnectionStateDisconnected {
			// Unregister tracks
			videoTracks = removeTrack(videoTracks, videoTrack)
			audioTracks = removeTrack(audioTracks, audioTrack)
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

// Serve WebRTC media streaming server
func Serve(remoteSdpChan, localSdpChan chan webrtc.SessionDescription, cfg *Options) {
	log.Printf("WebRTC server using UDP from %d to %d", cfg.MinPortUDP, cfg.MaxPortUDP)

	go playVideo()
	go playAudio()

	// Handle new connections
	for {
		// Wait for incoming session description
		// then send the local description to browser
		offer := <-remoteSdpChan
		localSdpChan <- newPeerHandler(offer, cfg)
	}
}
