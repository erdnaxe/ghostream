document.getElementsByName("quality").forEach(function (elem) {
    elem.onclick = function () {
        // Restart the connection with a new quality
        peerConnection.createOffer({"iceRestart": true}).then(offer => {
            return peerConnection.setLocalDescription(offer)
        }).catch(console.log)
    }
})
