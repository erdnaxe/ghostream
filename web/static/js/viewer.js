// Init peer connection
peerConnection = new RTCPeerConnection({
    iceServers: [
        {
            // FIXME: let admin customize the stun server
            urls: 'stun:stun.l.google.com:19302'
        }
    ]
})

// On connection change, inform user
peerConnection.oniceconnectionstatechange = e => {
    console.log(peerConnection.iceConnectionState)

    switch (myPeerConnection.iceConnectionState) {
        case "closed":
        case "failed":
            console.log("FIXME Failed");
            break;
        case "disconnected":
            console.log("temp network issue")
            break;
        case "connected":
            console.log("temp network issue resolved!")
            break;
    }
}

// We want to receive audio and video
peerConnection.addTransceiver('video', { 'direction': 'sendrecv' })
peerConnection.addTransceiver('audio', { 'direction': 'sendrecv' })

// Create offer and set local description
peerConnection.createOffer().then(offer => {
    // After setLocalDescription, the browser will fire onicecandidate events
    peerConnection.setLocalDescription(offer)
}).catch(console.log)

// When candidate is null, ICE layer has run out of potential configurations to suggest
// so let's send the offer to the server
peerConnection.onicecandidate = event => {
    if (event.candidate === null) {
        // Send offer to server
        // The server know the stream name from the url
        // The server replies with its description
        // After setRemoteDescription, the browser will fire ontrack events
        fetch(window.location.href, {
            method: 'POST',
            headers: {
                'Accept': 'application/json',
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(peerConnection.localDescription)
        })
            .then(response => response.json())
            .then((data) => peerConnection.setRemoteDescription(new RTCSessionDescription(data)))
            .catch(console.log)
    }
}

// When video track is received, configure player
peerConnection.ontrack = function (event) {
    if (event.track.kind === "video") {
        const viewer = document.getElementById('viewer')
        viewer.srcObject = event.streams[0]
        viewer.autoplay = true
        viewer.controls = true
        viewer.muted = true
    }
}
