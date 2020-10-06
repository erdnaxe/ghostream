document.getElementById("quality").addEventListener("change", (event) => {
    console.log(`Stream quality changed to ${event.target.value}`)

    // Restart the connection with a new quality
    // FIXME: set quality
    peerConnection.createOffer({ "iceRestart": true }).then(offer => {
        return peerConnection.setLocalDescription(offer)
    }).catch(console.log)
})
