package stream

import (
	"context"
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

const (
	audioFileName = "output.ogg"
	videoFileName = "output.ivf"
)

var (
	iceConnectedCtx, iceConnectedCtxCancel = context.WithCancel(context.Background())
)

// newPeerHandler is called when server receive a new session description
// this initiates a WebRTC connection and return server description
func newPeerHandler(api *webrtc.API, remoteSdp webrtc.SessionDescription, audioTrack *webrtc.Track, videoTrack *webrtc.Track) webrtc.SessionDescription {
	// Create a new PeerConnection
	peerConnection, err := api.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		log.Println("Failed to initiate peer connection", err)
		return webrtc.SessionDescription{}
	}

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		log.Printf("Connection State has changed %s \n", connectionState.String())
		if connectionState == webrtc.ICEConnectionStateConnected {
			iceConnectedCtxCancel()
		}
	})

	// Add audio and video tracks
	if _, err = peerConnection.AddTrack(audioTrack); err != nil {
		log.Println("Failed to add audio track", err)
		return webrtc.SessionDescription{}
	}
	if _, err = peerConnection.AddTrack(videoTrack); err != nil {
		log.Println("Failed to add video track", err)
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

	// Block until ICE Gathering is complete, disabling trickle ICE
	// we do this because we only can exchange one signaling message
	// in a production application you should exchange ICE Candidates via OnICECandidate
	<-gatherComplete

	// Output the local description and send it to browser
	return *peerConnection.LocalDescription()
}

// Serve WebRTC media streaming server
func Serve(remoteSdpChan chan webrtc.SessionDescription, localSdpChan chan webrtc.SessionDescription) {
	// Assert that we have an audio or video file
	_, err := os.Stat(videoFileName)
	haveVideoFile := !os.IsNotExist(err)
	_, err = os.Stat(audioFileName)
	haveAudioFile := !os.IsNotExist(err)
	if !haveAudioFile || !haveVideoFile {
		panic("Could not find `" + audioFileName + "` or `" + videoFileName + "`")
	}

	// Create media engine
	// Only support VP8 and Opus
	mediaEngine := webrtc.MediaEngine{}
	offer := <-remoteSdpChan
	if err = mediaEngine.PopulateFromSDP(offer); err != nil {
		panic(err)
	}

	// Create a new API object
	api := webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngine))

	// Create video track
	codec, payloadType := getPayloadType(mediaEngine, webrtc.RTPCodecTypeVideo, "VP8")
	videoTrack, err := webrtc.NewTrack(payloadType, rand.Uint32(), "video", "pion", codec)
	if err != nil {
		panic(err)
	}

	// Create audio track
	codec, payloadType = getPayloadType(mediaEngine, webrtc.RTPCodecTypeAudio, "opus")
	audioTrack, err := webrtc.NewTrack(payloadType, rand.Uint32(), "audio", "pion", codec)
	if err != nil {
		panic(err)
	}

	localSdpChan <- newPeerHandler(api, offer, audioTrack, videoTrack)

	go func() {
		// Open a IVF file and start reading using our IVFReader
		file, ivfErr := os.Open(videoFileName)
		if ivfErr != nil {
			panic(ivfErr)
		}

		ivf, header, ivfErr := ivfreader.NewWith(file)
		if ivfErr != nil {
			panic(ivfErr)
		}

		// Wait for connection established
		<-iceConnectedCtx.Done()

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
			if ivfErr = videoTrack.WriteSample(media.Sample{Data: frame, Samples: 90000}); ivfErr != nil {
				log.Fatalln("Failed to write video stream:", ivfErr)
			}
		}
	}()

	go func() {
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

		// Wait for connection established
		<-iceConnectedCtx.Done()

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

			if oggErr = audioTrack.WriteSample(media.Sample{Data: pageData, Samples: uint32(sampleCount)}); oggErr != nil {
				log.Fatalln("Failed to write audio stream:", oggErr)
			}

			// Convert seconds to Milliseconds, Sleep doesn't accept floats
			time.Sleep(time.Duration((sampleCount/48000)*1000) * time.Millisecond)
		}
	}()

	// Handle new connections
	for {
		// Wait for incoming session description
		// then send the local description to browser
		offer := <-remoteSdpChan
		localSdpChan <- newPeerHandler(api, offer, audioTrack, videoTrack)
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
