import React, { useContext, useEffect } from 'react';
import { ApiContext } from './contexts.mjs';

export const useApi = () => useContext(ApiContext)

export const useFeed = ({
    onConnect,
    onDisconnect,
    onError,
    onConnectStatus,
    onSessionCount,
    onMessage,
}) => {
    const api = useApi();
    useEffect(() => {
         
        const handleMessage = (e) => {
            const data = e.data && JSON.parse(e.data)
            switch (data.key) {
                case 'amazon-connect':
                    onConnectStatus && onConnectStatus(data.value);
                    break;
                case 'active-sessions':
                    onSessionCount && onSessionCount(data.value);
                    break;
                default:
                    onMessage && onMessage(e)
            }
        }

        api.connectFeed({ onMessage: handleMessage, onConnect, onDisconnect, onError })

    }, []);
};