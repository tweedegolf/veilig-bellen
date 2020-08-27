
// Regeister a new status feed listener
const registerFeedListener = (feedListeners) => ({ onConnect, onMessage, onDisconnect, onError }) => {
    feedListeners.push({ onConnect, onMessage, onDisconnect, onError });
}

// Remove a status feed listener
const removeFeedListener = (feedListeners) => (l) => feedListeners.remove(l)

// Initialize a automatically-reconnecting websocket connection
// to the agent status feed
const initWebsocket = (url, feedListeners) => {
    let state = { websocket: null, closed: false, reconnectInterval: null };

    const connect = () => {
        let { websocket } = state;
        const { closed } = state;

        // Cancel pending connection
        if (websocket !== null && !closed) {
            if (websocket.readyState === WebSocket.CONNECTING) {
                return; // Still trying to connect
            } else {
                websocket.onclose = () => console.log('Canceled connection attempt');
                websocket.close();
            }
        }
        // Try to reconnect
        console.log('Connecting to status feed...');
        websocket = new WebSocket(url);

        // Connection successful
        websocket.onopen = (e) => {
            const { reconnectInterval } = state;
            console.log('Connected to status feed')
            // Cancel reconnection interval
            if (reconnectInterval !== null) {
                clearInterval(reconnectInterval);
                state = { ...state, reconnectInterval: null };
            }
            feedListeners.forEach(({ onConnect }) => onConnect && onConnect(e));
        }

        // Message received
        websocket.onmessage = (e) => feedListeners.forEach(({ onMessage }) => onMessage && onMessage(e));

        // Connection closed, either by error or by websocket.close()
        websocket.onclose = (e) => {
            const { closed, reconnectInterval } = state;
            if (!closed) {
                // Set a reconnection interval
                if (reconnectInterval === null) {
                    state = { ...state, reconnectInterval: setInterval(connect, 1000) };
                }
                console.log('Disconnected from status feed, trying to reconnect...')
            }
            feedListeners.forEach(({ onDisconnect }) => onDisconnect && onDisconnect(e));
        };

        // An error happened
        websocket.onerror = (e) => feedListeners.forEach(({ onError }) => onError && onError(e))

        // Store websocket in state
        state = { ...state, websocket };
    }

    connect();

    // Call this method to close the feed and disable reconnecting
    const close = () => {
        state.closed = true;
        if (state.websocket) {
            state.websocket.close();
        }
        console.log("Closing feed...");
    }

    return close;
}

// Initialize the feed connection.
export const initFeed = (feedUrl) => {
    const feedListeners = [];

    const closeFeed = initWebsocket(feedUrl, feedListeners);

    return {
        registerFeedListener: registerFeedListener(feedListeners),
        removeFeedListener: removeFeedListener(feedListeners),
        closeFeed,
    }
}