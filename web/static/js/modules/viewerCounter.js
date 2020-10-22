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
        fetch(this.url)
            .then(response => response.json())
            .then((data) => this.element.innerText = data.ConnectedViewers)
            .catch(console.log);
    }
}
