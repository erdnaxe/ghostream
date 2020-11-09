import { ViewerCounter } from "./modules/viewerCounter.js";

/**
 * Initialize viewer page
 *
 * @param {String} stream
 * @param {Number} viewersCounterRefreshPeriod
 * @param {String} posterUrl
 */
export function initViewerPage(stream, viewersCounterRefreshPeriod, posterUrl) {
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

    // Create player
    let player = OvenPlayer.create("viewer", {
        title: stream,
        image: posterUrl,
        autoStart: true,
        mute: true,
        expandFullScreenUI: true,
        sources: [
            {
                "file": "wss://" + window.location.host + "/app/{{.Path}}",
                "type": "webrtc",
                "label": " WebRTC - Source"
            },
            {
                "type": "hls",
                "file": "https://" + window.location.host + "/app/{{.Path}}_bypass/playlist.m3u8",
                "label": " HLS"
            }
        ]
    });
    player.on("error", function (error) {
        if (error.code === 501 || error.code === 406) {
            // Clear messages
            const errorMsg = document.getElementsByClassName("op-message-text")[0]
            errorMsg.textContent = ""

            const warningIcon = document.getElementsByClassName("op-message-icon")[0]
            warningIcon.textContent = ""

            // Reload in 30s
            setTimeout(function () {
                player.load()
            }, 30000)
        } else {
            console.log(error);
        }
    });
}
