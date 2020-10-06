document.getElementById("quality").addEventListener("change", (event) => {
    console.log(`Stream quality changed to ${event.target.value}`)

    // Restart the connection with a new quality
    peerConnection.close()
    peerConnection = null
    streamPath = window.location.href + event.target.value
    startPeerConnection()
})
