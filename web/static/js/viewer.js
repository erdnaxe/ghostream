// Init peer connection
let pc = new RTCPeerConnection({
    iceServers: [
        {
            // FIXME: let admin customize the stun server
            urls: 'stun:stun.l.google.com:19302'
        }
    ]
})

// Create an offer to receive one video and one audio track
pc.addTransceiver('video', { 'direction': 'sendrecv' })
pc.addTransceiver('audio', { 'direction': 'sendrecv' })
pc.createOffer().then(d => pc.setLocalDescription(d)).catch(console.log)

// When local session description is ready, send it to streaming server
// FIXME: also send stream path
// FIXME: send to wss://{{.Cfg.Hostname}}/play/{{.Path}}
pc.oniceconnectionstatechange = e => console.log(pc.iceConnectionState)
pc.onicecandidate = event => {
    if (event.candidate === null) {
        document.getElementById('localSessionDescription').value = JSON.stringify(pc.localDescription)
    }
}

// When remote session description is received, load it
window.startSession = () => {
    let sd = document.getElementById('remoteSessionDescription').value
    try {
        pc.setRemoteDescription(new RTCSessionDescription(JSON.parse(sd)))
    } catch (e) {
        console.log(e)
    }
}

// When video track is received, mount player
pc.ontrack = function (event) {
    if (event.track.kind === "video") {
        const viewer = document.getElementById('viewer')
        viewer.srcObject = event.streams[0]
        viewer.autoplay = true
        viewer.controls = true
        viewer.muted = true
    }
}
