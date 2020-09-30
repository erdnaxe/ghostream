// Refresh viewer count by pulling metric from server
function refreshViewersCounter(period) {
    fetch("/_stats/")
        .then(response => response.json())
        .then((data) => document.getElementById("connected-people").innerText = data.ConnectedViewers)
        .catch(console.log)

    setTimeout(() => {
        refreshViewersCounter(period)
    }, period)
}
