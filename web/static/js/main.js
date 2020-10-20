import { GSWebSocket } from "./modules/websocket.js";
import { ViewerCounter } from "./modules/viewerCounter.js";

// Create WebSocket
const s = new GSWebSocket();
s.open(() => {
    // FIXME open callback
}, () => {
    // FIXME close callback
});

// Create viewer counter
const streamName = "demo"; // FIXME
const viewerCounter = new ViewerCounter(
    document.getElementById("connected-people"),
    streamName,
);
viewerCounter.regularUpdate(1000);  // FIXME
