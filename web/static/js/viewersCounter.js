// Refresh viewer count by pulling metric from server
function refreshViewersCounter(period) {
    // Distinguish oneDomainPerStream mode
    fetch("/_stats/" + (location.pathname === "/" ? location.host : location.pathname.substring(1)))
        .then(response => response.json())
        .then((data) => document.getElementById("connected-people").innerText = data.ConnectedViewers)
        .catch(console.log)

    setTimeout(() => {
        refreshViewersCounter(period)
    }, period)
}
