
// Regeister a new status feed listener
const registerFeedListener = (feedListeners) => ({ onConnect, onMessage, onDisconnect, onError }) => {
    feedListeners.push({ onConnect, onMessage, onDisconnect, onError });
}

// Remove a status feed listener
const removeFeedListener = (feedListeners) => (l) => feedListeners.remove(l)

// Initialize a automatically-reconnecting websocket connection
// to the agent status feed
const initFeed = (backendHostname, feedListeners) => {
    let reconnectInterval = null
    const connect = () => {
        console.log('Connecting to status feed...');
        const websocket = new WebSocket(`wss://${backendHostname}/agent-feed`);

        websocket.onopen = (e) => {
            console.log('Connected to status feed')
            if(reconnectInterval !== null) {
                clearInterval(reconnectInterval);
                reconnectInterval = null;
            }
            feedListeners.forEach(({ onConnect }) => onConnect && onConnect(e));
        }

        websocket.onmessage = (e) => feedListeners.forEach(({ onMessage }) => onMessage && onMessage(e));
        websocket.onclose = (e) => {
            if(reconnectInterval === null) {
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
export const initApi = () => {
    const backendHostname = process.env.BACKEND_HOSTNAME;
    const feedListeners = [];

    initFeed(backendHostname, feedListeners);

    return {
        registerFeedListener: registerFeedListener(feedListeners),
        removeFeedListener: removeFeedListener(feedListeners),
    }
}