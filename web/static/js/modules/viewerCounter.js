/**
 * ViewerCounter show the number of active viewers
 */
export class ViewerCounter {
    /**
     * @param {HTMLElement} element 
     * @param {String} streamName 
     */
    constructor(element, streamName) {
        this.element = element;
        this.url = "/_stats/" + streamName;
        this.uid = Math.floor(1e19 * Math.random()).toString(16);
    }

    /**
     * Regulary update counter
     * 
     * @param {Number} updatePeriod 
     */
    regularUpdate(updatePeriod) {
        setInterval(() => this.refreshViewersCounter(), updatePeriod);
    }

    refreshViewersCounter() {
        fetch(this.url + "?uid=" + this.uid)
            .then(response => response.json())
            .then((data) => this.element.innerText = data.ConnectedViewers)
            .catch(console.log);
    }
}
