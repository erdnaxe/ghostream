// Package webrtc provides the backend to simulate a WebRTC client to send stream
package webrtc

import (
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
func newPeerHandler(localSdpChan chan webrtc.SessionDescription, remoteSdp struct {
	StreamID          string
	RemoteDescription webrtc.SessionDescription
}, cfg *Options) {
	// Create media engine using client SDP
	mediaEngine := webrtc.MediaEngine{}
	if err := mediaEngine.PopulateFromSDP(remoteSdp.RemoteDescription); err != nil {
		log.Println("Failed to create new media engine", err)
		localSdpChan <- webrtc.SessionDescription{}
		return
	}

	// Create a new PeerConnection
	settingsEngine := webrtc.SettingEngine{}
	if err := settingsEngine.SetEphemeralUDPPortRange(cfg.MinPortUDP, cfg.MaxPortUDP); err != nil {
		log.Println("Failed to set min/max UDP ports", err)
		localSdpChan <- webrtc.SessionDescription{}
		return
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
		localSdpChan <- webrtc.SessionDescription{}
		return
	}

	// Create video track
	codec, payloadType := getPayloadType(mediaEngine, webrtc.RTPCodecTypeVideo, "VP8")
	videoTrack, err := webrtc.NewTrack(payloadType, rand.Uint32(), "video", "pion", codec)
	if err != nil {
		log.Println("Failed to create new video track", err)
		localSdpChan <- webrtc.SessionDescription{}
		return
	}
	if _, err = peerConnection.AddTrack(videoTrack); err != nil {
		log.Println("Failed to add video track", err)
		localSdpChan <- webrtc.SessionDescription{}
		return
	}

	// Create audio track
	codec, payloadType = getPayloadType(mediaEngine, webrtc.RTPCodecTypeAudio, "opus")
	audioTrack, err := webrtc.NewTrack(payloadType, rand.Uint32(), "audio", "pion", codec)
	if err != nil {
		log.Println("Failed to create new audio track", err)
		localSdpChan <- webrtc.SessionDescription{}
		return
	}
	if _, err = peerConnection.AddTrack(audioTrack); err != nil {
		log.Println("Failed to add audio track", err)
		localSdpChan <- webrtc.SessionDescription{}
		return
	}

	// Set the remote SessionDescription
	if err = peerConnection.SetRemoteDescription(remoteSdp.RemoteDescription); err != nil {
		log.Println("Failed to set remote description", err)
		localSdpChan <- webrtc.SessionDescription{}
		return
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

	// Create answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		log.Println("Failed to create answer", err)
		localSdpChan <- webrtc.SessionDescription{}
		return
	}

	// Sets the LocalDescription
	// Using GatheringCompletePromise disable trickle ICE
	// FIXME: https://github.com/pion/webrtc/wiki/Release-WebRTC@v3.0.0
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	if err = peerConnection.SetLocalDescription(answer); err != nil {
		log.Println("Failed to set local description", err)
		localSdpChan <- webrtc.SessionDescription{}
		return
	}
	<-gatherComplete

	// Send answer to client
	localSdpChan <- *peerConnection.LocalDescription()
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
	log.Printf("Remote peer does not support %s", codecName)
	return nil, 0
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
		newPeerHandler(localSdpChan, <-remoteSdpChan, cfg)
	}
}
