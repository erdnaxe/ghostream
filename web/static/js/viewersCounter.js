function refreshViewersCounter() {
    let xhr = new XMLHttpRequest()
    xhr.open("GET", "/_stats/", true)
    xhr.onload = function () {
        console.log(xhr.response)
        if (xhr.status === 200) {
            let data = JSON.parse(xhr.response)
            document.getElementById("connected-people").innerText = data.ConnectedViewers
        }
        else
            console.log("WARNING: status code " + xhr.status + " was returned while fetching connected viewers.")
    }
    xhr.send()

    setTimeout(refreshViewersCounter, 20000)
}

refreshViewersCounter()
