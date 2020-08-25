
// Regeister a new status feed listener
const registerFeedListener = (feedListeners) => ({ onConnect, onMessage, onDisconnect, onError }) => {
    feedListeners.push({ onConnect, onMessage, onDisconnect, onError });
}

// Remove a status feed listener
const removeFeedListener = (feedListeners) => (l) => feedListeners.remove(l)

// Initialize a automatically-reconnecting websocket connection
// to the agent status feed
const initFeed = (backendFeedUrl, feedListeners) => {
    let reconnectInterval = null;
    let websocket = null;
    
    const connect = () => {
        // Cancel pending connection
        if (websocket !== null) {
            if (websocket.readyState === WebSocket.CONNECTING) {
                return; // Still trying to connect
            } else {
                websocket.onclose = () => console.log('Canceled connection attempt');
                websocket.close();
            }
        }
        // Try to reconnect
        console.log('Connecting to status feed...');
        websocket = new WebSocket(`${backendFeedUrl}`);

        websocket.onopen = (e) => {
            console.log('Connected to status feed')
            // Cancel reconnection interval
            if (reconnectInterval !== null) {
                clearInterval(reconnectInterval);
                reconnectInterval = null;
            }
            feedListeners.forEach(({ onConnect }) => onConnect && onConnect(e));
        }

        websocket.onmessage = (e) => feedListeners.forEach(({ onMessage }) => onMessage && onMessage(e));
        websocket.onclose = (e) => {
            // Set a reconnection interval
            if (reconnectInterval === null) {
                reconnectInterval = setInterval(connect, 1000);
            }
            console.log('Disconnected from status feed, tring to reconnect...')
            feedListeners.forEach(({ onDisconnect }) => onDisconnect && onDisconnect(e));
        };
        websocket.onerror = (e) => feedListeners.forEach(({ onError }) => onError && onError(e))
    }

    connect()

}

// Initialize the Api connections, setting up the feed connection.
export const initApi = (backendFeedUrl) => {
    const feedListeners = [];

    initFeed(backendFeedUrl, feedListeners);

    return {
        registerFeedListener: registerFeedListener(feedListeners),
        removeFeedListener: removeFeedListener(feedListeners),
    }
}