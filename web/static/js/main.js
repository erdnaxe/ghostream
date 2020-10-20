import { GSWebSocket } from "./modules/websocket.js";
import { ViewerCounter } from "./modules/viewerCounter.js";

// Some variables that need to be fixed by web page
const ViewersCounterRefreshPeriod = 1000;
const stream = "demo";
let quality = "source";

// Create WebSocket
const s = new GSWebSocket();
s.open();

// Register keyboard events
const viewer = document.getElementById("viewer");
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
viewerCounter.regularUpdate(ViewersCounterRefreshPeriod);

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

    // Restart the connection with a new quality
    // FIXME
});
