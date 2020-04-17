

const connectFeed = (backendHostname) => async ({ onConnect, onMessage, onDisconnect, onError }) => {
    const websocket = new WebSocket(`wss://${backendHostname}/agent-feed`);

    websocket.onopen = (e) => onConnect && onConnect(e);
    websocket.onmessage = (e) => onMessage && onMessage(e);
    websocket.onclose = (e) => onDisconnect && onDisconnect(e);
    websocket.onerror = (e) => onError && onError(e);
}

export const initApi = () => {
    const backendHostname = process.env.BACKEND_HOSTNAME;

    return {
        connectFeed: connectFeed(backendHostname)
    }
}