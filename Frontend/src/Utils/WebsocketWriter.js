class WebSocketWriter {
    constructor(URL, accessTokenRef, onclose) {
        this.ws = new WebSocket(URL);
        this.lastUsed = Date.now();
        this.queue = [];
        this.locked = false;
        this.accessTokenRef = accessTokenRef;
        this.authRequired = !!accessTokenRef;
        this.openPromise = new Promise((resolve, reject) => {
            this.ws.onerror = reject;
            this.ws.onclose = onclose;
            this.ws.onopen = () => {
                this.authenticateIfRequired()
                resolve(this)
            };
        });
    }

    send(reason, message) {
        if (this.state() === this.ws.readyState) {
            this.queue.push(JSON.stringify({reason: reason, message: message}));
            this.flush();
            return true;
        }
        return false;
    }

    state() {
        return this.ws.readyState;
    }

    flush() {
        if (this.locked) return;
        if (!this.ws || this.ws.readyState !== WebSocket.OPEN) return;
        if (this.queue.length === 0) return;

        this.locked = true;

        try {
            const msg = this.queue.shift();
            this.ws.send(msg);
            this.lastUsed = Date.now();
        } finally {
            this.locked = false;
            queueMicrotask(() => this.flush());
        }
    }

    authenticateIfRequired() {
        if (this.authRequired) this.send("auth", this.accessTokenRef.current)
    }
}
