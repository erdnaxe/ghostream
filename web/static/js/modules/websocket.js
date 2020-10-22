/**
 * GsWebSocket to do Ghostream signalling
 */
export class GsWebSocket {
    constructor() {
        const protocol = (window.location.protocol === "https:") ? "wss://" : "ws://";
        this.url = protocol + window.location.host + "/_ws/";

        // Open WebSocket
        this._open();

        // Configure events
        this.socket.addEventListener("open", () => {
            console.log("[WebSocket] Connection established");
        });
        this.socket.addEventListener("close", () => {
            console.log("[WebSocket] Connection closed, retrying connection in 1s...");
            setTimeout(() => this._open(), 1000);
        });
        this.socket.addEventListener("error", () => {
            console.log("[WebSocket] Connection errored, retrying connection in 1s...");
            setTimeout(() => this._open(), 1000);
        });
    }

    _open() {
        console.log(`[WebSocket] Connecting to ${this.url}...`);
        this.socket = new WebSocket(this.url);
    }

    /**
     * Send local WebRTC session description to remote.
     * @param {SessionDescription} localDescription WebRTC local SDP
     * @param {string} stream Name of the stream
     * @param {string} quality Requested quality 
     */
    sendLocalDescription(localDescription, stream, quality) {
        if (this.socket.readyState !== 1) {
            console.log("[WebSocket] Waiting for connection to send data...");
            setTimeout(() => this.sendDescription(localDescription, stream, quality), 100);
            return;
        }
        console.log(`[WebSocket] Sending WebRTC local session description for stream ${stream} quality ${quality}`);
        this.socket.send(JSON.stringify({
            "webRtcSdp": localDescription,
            "stream": stream,
            "quality": quality
        }));
    }

    /**
     * Set callback function on new remote session description.
     * @param {Function} callback Function called when data is received
     */
    onRemoteDescription(callback) {
        this.socket.addEventListener("message", (event) => {
            console.log("[WebSocket] Received WebRTC remote session description");
            const sdp = new RTCSessionDescription(JSON.parse(event.data));
            callback(sdp);
        });
    }
}
