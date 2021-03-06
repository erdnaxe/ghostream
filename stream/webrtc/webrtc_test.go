package webrtc

import (
	"math/rand"
	"testing"

	"github.com/pion/webrtc/v3"
	"gitlab.crans.org/nounous/ghostream/messaging"
)

func TestServe(t *testing.T) {
	// Init streams messaging and WebRTC server
	streams := messaging.New()
	cfg := Options{
		Enabled:     true,
		MinPortUDP:  10000,
		MaxPortUDP:  10005,
		STUNServers: []string{"stun:stun.l.google.com:19302"},
	}
	go Serve(streams, &cfg)

	// New client connection
	mediaEngine := webrtc.MediaEngine{}
	mediaEngine.RegisterDefaultCodecs()
	api := webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngine))
	peerConnection, _ := api.NewPeerConnection(webrtc.Configuration{})

	// Create video track
	codec, payloadType := getPayloadType(mediaEngine, webrtc.RTPCodecTypeVideo, "H264")
	videoTrack, err := webrtc.NewTrack(payloadType, rand.Uint32(), "video", "pion", codec)
	if err != nil {
		t.Error("Failed to create new video track", err)
	}
	if _, err = peerConnection.AddTrack(videoTrack); err != nil {
		t.Error("Failed to add video track", err)
	}

	// Create audio track
	codec, payloadType = getPayloadType(mediaEngine, webrtc.RTPCodecTypeAudio, "opus")
	audioTrack, err := webrtc.NewTrack(payloadType, rand.Uint32(), "audio", "pion", codec)
	if err != nil {
		t.Error("Failed to create new audio track", err)
	}
	if _, err = peerConnection.AddTrack(audioTrack); err != nil {
		t.Error("Failed to add audio track", err)
	}

	// Create offer
	offer, _ := peerConnection.CreateOffer(nil)

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	peerConnection.SetLocalDescription(offer)
	<-gatherComplete

	// FIXME: Send offer to server
	// FIXME: verify connection did work
}
