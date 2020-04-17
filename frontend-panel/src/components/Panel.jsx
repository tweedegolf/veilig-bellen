import React, { useState, useEffect } from 'react';
import { useApi, useFeed } from "../hooks";

const handleConnectStatus = (setConnectStatus) => (status) => {

    const update = {};
    update.lastUpdate = new Date(status.DataSnapshotTime);
    status.MetricResults.forEach(r => r.Collections.forEach(({ Metric: { Name: n }, Value: v }) => {
        switch (n) {
            case 'AGENTS_ONLINE':
                update.agentsOnline = v
                break;
            case 'AGENTS_AVAILABLE':
                update.agentsAvailable = v;
                break;
            case 'AGENTS_ON_CALL':
                update.agentsOnCall = v;
                break;
            case 'CONTACTS_IN_QUEUE':
                update.contactsInQueue = v;
                break;
        }
    }));

    setConnectStatus(update);
}

const Panel = () => {
    const [state, setState] = useState({
        connected: false,
        message: null,
        connectStatus: {
            lastUpdate: null,
            agentsOnline: null,
            agentsAvailable: null,
            agentsOnCall: null,
            contactsInQueue: null,
        },
        sessionCount: null,
        error: null,
    });

    useFeed({
        onConnect: () => setState(s => ({ ...s, connected: true, error: false })),
        onMessage: e => console.log('Unrecognized message', e),
        onDisconnect: () => setState(s => ({ ...s, connected: false })),
        onSessionCount: d => setState(s => ({ ...s, sessionCount: d.count })),
        onConnectStatus: handleConnectStatus(u => setState(s => ({ ...s, connectStatus: { ...s.connectStatus, ...u } }))),
        onError: e => setState(s => ({ ...s, error: e }))
    });

    if (!state.connected) {
        return (<p>Connecting...</p>)
    }

    const ConnectStatus = ({ s }) => (
        <>
            <tr>
                <th>Last Connect update</th>
                <td>{s.lastUpdate ? s.lastUpdate.toLocaleDateString() : '-'}</td>
            </tr>
            <tr>
                <th>Agents online</th>
                <td>{s.agentsOnline}</td>
            </tr>
            <tr>
                <th>Agents available</th>
                <td>{s.agentsAvailable}</td>
            </tr>
            <tr>
                <th>Agents on call</th>
                <td>{s.agentsOnCall}</td>
            </tr>
            <tr>
                <th>Contacts in queue</th>
                <td>{s.contactsInQueue}</td>
            </tr>
        </>);

    return (
        <div className="status-panel">
            <h1>Status panel</h1>
            <table>
                <tbody>
                    {state.connectStatus && <ConnectStatus s={state.connectStatus} />}
                    <tr>
                        <th>Active Irma sessions</th>
                        <td>{state.sessionCount}</td>
                    </tr>
                </tbody>
            </table>
        </div>
    )
};


export default Panel;