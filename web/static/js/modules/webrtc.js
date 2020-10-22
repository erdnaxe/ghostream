/**
 * GsWebRTC to connect to Ghostream
 */
export class GsWebRTC {
    /**
     * @param {list} stunServers STUN servers
     * @param {HTMLElement} viewer Video HTML element
     * @param {HTMLElement} connectionIndicator Connection indicator element
     */
    constructor(stunServers, viewer, connectionIndicator) {
        this.viewer = viewer;
        this.connectionIndicator = connectionIndicator;
        this.pc = new RTCPeerConnection({
            iceServers: [{ urls: stunServers }]
        });

        // We want to receive audio and video
        this.pc.addTransceiver("video", { "direction": "sendrecv" });
        this.pc.addTransceiver("audio", { "direction": "sendrecv" });

        // Configure events
        this.pc.oniceconnectionstatechange = () => this._onConnectionStateChange();
        this.pc.ontrack = (e) => this._onTrack(e);
    }

    /**
     * On connection change, log it and change indicator.
     * If connection closed or failed, try to reconnect.
     */
    _onConnectionStateChange() {
        console.log("[WebRTC] ICE connection state changed to " + this.pc.iceConnectionState);
        switch (this.pc.iceConnectionState) {
        case "disconnected":
            this.connectionIndicator.style.fill = "#dc3545";
            break;
        case "checking":
            this.connectionIndicator.style.fill = "#ffc107";
            break;
        case "connected":
            this.connectionIndicator.style.fill = "#28a745";
            break;
        case "closed":
        case "failed":
            console.log("[WebRTC] Connection closed, restarting...");
            /*peerConnection.close();
                    peerConnection = null;
                    setTimeout(startPeerConnection, 1000);*/
            break;
        }
    }

    /**
     * On new track, add it to the player
     * @param {Event} event 
     */
    _onTrack(event) {
        console.log(`[WebRTC] New ${event.track.kind} track`);
        if (event.track.kind === "video") {
            this.viewer.srcObject = event.streams[0];
        }
    }

    /**
     * Create an offer and set local description.
     * After that the browser will fire onicecandidate events.
     */
    createOffer() {
        this.pc.createOffer().then(offer => {
            this.pc.setLocalDescription(offer);
            console.log("[WebRTC] WebRTC offer created");
        }).catch(console.log);
    }

    /**
     * Register a function to call to send local descriptions
     * @param {Function} sendFunction Called with a local description to send.
     */
    onICECandidate(sendFunction) {
        // When candidate is null, ICE layer has run out of potential configurations to suggest
        // so let's send the offer to the server.
        // FIXME: Send offers progressively to do Trickle ICE
        this.pc.onicecandidate = event => {
            if (event.candidate === null) {
                // Send offer to server
                console.log("[WebRTC] Sending session description to server");
                sendFunction(this.pc.localDescription);
            }
        };
    }

    /**
     * Set WebRTC remote description
     * After that, the connection will be established and ontrack will be fired.
     * @param {RTCSessionDescription} sdp Session description data
     */
    setRemoteDescription(sdp) {
        this.pc.setRemoteDescription(sdp);
    }
}
