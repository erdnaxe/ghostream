import { GsWebSocket } from "./modules/websocket.js";
import { ViewerCounter } from "./modules/viewerCounter.js";
import { GsWebRTC } from "./modules/webrtc.js";

/**
 * Initialize viewer page
 * 
 * @param {String} stream 
 * @param {List} stunServers 
 * @param {Number} viewersCounterRefreshPeriod 
 */
export function initViewerPage(stream, stunServers, viewersCounterRefreshPeriod) {
    // Viewer element
    const viewer = document.getElementById("viewer");

    // Default quality
    let quality = "source";

    // Create WebSocket and WebRTC
    const websocket = new GsWebSocket();
    const webrtc = new GsWebRTC(
        stunServers,
        viewer,
        document.getElementById("connectionIndicator"),
    );
    webrtc.createOffer();
    webrtc.onICECandidate(localDescription => {
        websocket.sendLocalDescription(localDescription, stream, quality);
    });
    websocket.onRemoteDescription(sdp => {
        webrtc.setRemoteDescription(sdp);
    });

    // Register keyboard events
    window.addEventListener("keydown", (event) => {
        switch (event.key) {
        case "f":
            // F key put player in fullscreen
            if (document.fullscreenElement !== null) {
                document.exitFullscreen();
            } else {
                viewer.requestFullscreen();
            }
            break;
        case "m":
        case " ":
            // M and space key mute player
            viewer.muted = !viewer.muted;
            event.preventDefault();
            viewer.play();
            break;
        }
    });

    // Create viewer counter
    const viewerCounter = new ViewerCounter(
        document.getElementById("connected-people"),
        stream,
    );
    viewerCounter.regularUpdate(viewersCounterRefreshPeriod);
    viewerCounter.refreshViewersCounter();

    // Side widget toggler
    const sideWidgetToggle = document.getElementById("sideWidgetToggle");
    const sideWidget = document.getElementById("sideWidget");
    if (sideWidgetToggle !== null && sideWidget !== null) {
        // On click, toggle side widget visibility
        sideWidgetToggle.addEventListener("click", function () {
            if (sideWidget.style.display === "none") {
                sideWidget.style.display = "block";
                sideWidgetToggle.textContent = "»";
            } else {
                sideWidget.style.display = "none";
                sideWidgetToggle.textContent = "«";
            }
        });
    }

    // Video quality toggler
    document.getElementById("quality").addEventListener("change", (event) => {
        quality = event.target.value;
        console.log(`Stream quality changed to ${quality}`);

        // Restart WebRTC negociation
        webrtc.createOffer();
    });
}
