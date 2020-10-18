let peerConnection
let streamPath = window.location.href

startPeerConnection = () => {
    // Init peer connection
    peerConnection = new RTCPeerConnection({
        iceServers: [{ urls: stunServers }]
    })

    // On connection change, change indicator color
    // if connection failed, restart peer connection
    peerConnection.oniceconnectionstatechange = e => {
        console.log("ICE connection state changed, " + peerConnection.iceConnectionState)
        switch (peerConnection.iceConnectionState) {
            case "disconnected":
                document.getElementById("connectionIndicator").style.fill = "#dc3545"
                break
            case "checking":
                document.getElementById("connectionIndicator").style.fill = "#ffc107"
                break
            case "connected":
                document.getElementById("connectionIndicator").style.fill = "#28a745"
                break
            case "closed":
            case "failed":
                console.log("Connection failed, restarting...")
                peerConnection.close()
                peerConnection = null
                setTimeout(startPeerConnection, 1000)
                break
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
            console.log("Sending session description to server")
            fetch(streamPath, {
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
        console.log(`New ${event.track.kind} track`)
        if (event.track.kind === "video") {
            const viewer = document.getElementById('viewer')
            viewer.srcObject = event.streams[0]
        }
    }
}

// Register keyboard events
let viewer = document.getElementById("viewer")
window.addEventListener("keydown", (event) => {
    switch (event.key) {
        case 'f':
            // F key put player in fullscreen
            if (document.fullscreenElement !== null) {
                document.exitFullscreen()
            } else {
                viewer.requestFullscreen()
            }
            break
        case 'm':
        case ' ':
            // M and space key mute player
            viewer.muted = !viewer.muted
            event.preventDefault()
            viewer.play()
            break
    }
})
