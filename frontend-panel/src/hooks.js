import React, { useContext, useEffect, useState } from 'react';
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
    const connect = () => api.registerFeedListener({ onMessage: handleMessage, onConnect, onDisconnect, onError });
    useEffect(() => {
         connect();
    }, []);
};