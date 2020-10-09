// Package webrtc provides the backend to simulate a WebRTC client to send stream
package webrtc

import (
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/pion/webrtc/v3"
	"gitlab.crans.org/nounous/ghostream/internal/monitoring"
	"gitlab.crans.org/nounous/ghostream/stream/srt"
)

// Options holds web package configuration
type Options struct {
	Enabled     bool
	MinPortUDP  uint16
	MaxPortUDP  uint16
	STUNServers []string
}

// SessionDescription contains SDP data
// to initiate a WebRTC connection between one client and this app
type SessionDescription = webrtc.SessionDescription

var (
	videoTracks map[string][]*webrtc.Track
	audioTracks map[string][]*webrtc.Track
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
func GetNumberConnectedSessions(streamID string) int {
	return len(videoTracks[streamID])
}

// newPeerHandler is called when server receive a new session description
// this initiates a WebRTC connection and return server description
func newPeerHandler(remoteSdp struct {
	StreamID          string
	RemoteDescription webrtc.SessionDescription
}, cfg *Options) webrtc.SessionDescription {
	// Create media engine using client SDP
	mediaEngine := webrtc.MediaEngine{}
	if err := mediaEngine.PopulateFromSDP(remoteSdp.RemoteDescription); err != nil {
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
	if err = peerConnection.SetRemoteDescription(remoteSdp.RemoteDescription); err != nil {
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

	streamID := remoteSdp.StreamID
	split := strings.SplitN(streamID, "@", 2)
	streamID = split[0]
	quality := "source"
	if len(split) == 2 {
		quality = split[1]
	}
	log.Printf("New WebRTC session for stream %s, quality %s", streamID, quality)
	// TODO Consider the quality

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		log.Printf("Connection State has changed %s \n", connectionState.String())
		if videoTracks[streamID] == nil {
			videoTracks[streamID] = make([]*webrtc.Track, 0, 1)
		}
		if audioTracks[streamID] == nil {
			audioTracks[streamID] = make([]*webrtc.Track, 0, 1)
		}
		if connectionState == webrtc.ICEConnectionStateConnected {
			// Register tracks
			videoTracks[streamID] = append(videoTracks[streamID], videoTrack)
			audioTracks[streamID] = append(audioTracks[streamID], audioTrack)
			monitoring.WebRTCConnectedSessions.Inc()
		} else if connectionState == webrtc.ICEConnectionStateDisconnected {
			// Unregister tracks
			videoTracks[streamID] = removeTrack(videoTracks[streamID], videoTrack)
			audioTracks[streamID] = removeTrack(audioTracks[streamID], audioTrack)
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
func Serve(remoteSdpChan chan struct {
	StreamID          string
	RemoteDescription webrtc.SessionDescription
}, localSdpChan chan webrtc.SessionDescription, inputChannel chan srt.Packet, cfg *Options) {
	if !cfg.Enabled {
		// SRT is not enabled, ignore
		return
	}

	log.Printf("WebRTC server using UDP from port %d to %d", cfg.MinPortUDP, cfg.MaxPortUDP)

	// Allocate memory
	videoTracks = make(map[string][]*webrtc.Track)
	audioTracks = make(map[string][]*webrtc.Track)

	// Ingest data from SRT
	go ingestFrom(inputChannel)

	// Handle new connections
	for {
		// Wait for incoming session description
		// then send the local description to browser
		localSdpChan <- newPeerHandler(<-remoteSdpChan, cfg)
	}
}
