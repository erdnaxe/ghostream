// Side widget toggler
const sideWidgetToggle = document.getElementById("sideWidgetToggle")
sideWidgetToggle.addEventListener("click", function () {
    const sideWidget = document.getElementById("sideWidget")
    if (sideWidget.style.display === "none") {
        sideWidget.style.display = "block"
        sideWidgetToggle.textContent = "»"
    } else {
        sideWidget.style.display = "none"
        sideWidgetToggle.textContent = "«"
    }
})