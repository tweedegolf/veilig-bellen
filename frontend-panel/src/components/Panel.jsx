import React, { useState, useEffect } from 'react';
import { useApi, useFeed } from "../hooks";

const Panel = () => {
    const [state, setState] = useState({
        connected: false,
        message: null,
        connectStatus: null,
        sessionCount: null,
        error: null,
    });

    useFeed({
        onConnect: () => setState(s => ({ ...s, connected: true })),
        onMessage: e => console.log(e),
        onDisconnect: () => setState(s => ({ ...s, connected: false })),
        onSessionCount: d => setState(s => ({...s, sessionCount: d.count})),
        onConnectStatus: status => setState(s => ({...s, connectStatus: status})),
    });

    if (!state.connected) {
        return (<p>Connecting... {JSON.stringify(state)}</p>)
    }

    return (<p>Panel {JSON.stringify(state)}</p>)
};


export default Panel;