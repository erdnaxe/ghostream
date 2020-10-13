// Refresh viewer count by pulling metric from server
function refreshViewersCounter(streamID, period) {
    // Distinguish oneDomainPerStream mode
    fetch("/_stats/" + streamID)
        .then(response => response.json())
        .then((data) => document.getElementById("connected-people").innerText = data.ConnectedViewers)
        .catch(console.log)

    setTimeout(() => {
        refreshViewersCounter(period)
    }, period)
}
