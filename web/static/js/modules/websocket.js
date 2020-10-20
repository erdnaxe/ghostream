/**
 * GsWebSocket to do Ghostream signalling
 */
export class GsWebSocket {
    constructor() {
        const protocol = (window.location.protocol === "https:") ? "wss://" : "ws://";
        this.url = protocol + window.location.host + "/_ws/";
    }

    _open() {
        this.socket = new WebSocket(this.url);
    }

    /**
     * Open websocket.
     * 
     * @param {Function} openCallback Function called when connection is established. 
     * @param {Function} closeCallback Function called when connection is lost. 
     */
    open() {
        this._open();
        this.socket.addEventListener("open", () => {
            console.log("WebSocket opened");
        });
        this.socket.addEventListener("close", () => {
            console.log("WebSocket closed, retrying connection in 1s...");
            setTimeout(this._open, 1000);
        });
        this.socket.addEventListener("error", () => {
            console.log("WebSocket errored, retrying connection in 1s...");
            setTimeout(this._open, 1000);
        });
    }

    /**
     * Exchange WebRTC session description with server.
     * 
     * @param {string} data JSON formated data 
     * @param {Function} receiveCallback Function called when data is received
     */
    exchangeDescription(data, receiveCallback) {
        if (this.socket.readyState !== 1) {
            console.log("WebSocket not ready to send data");
            return;
        }
        this.socket.send(data);
        this.socket.addEventListener("message", (event) => {
            console.log("Message from server ", event.data);
            receiveCallback(event);
        });
    }
}
