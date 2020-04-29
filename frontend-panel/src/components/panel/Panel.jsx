import React, { useState } from 'react';
import { useFeed } from '../../hooks';
import { handleConnectStatus } from '../../util';
import { Box, Container, Typography, Grid } from '@material-ui/core';
import PanelItem from './PanelItem';

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

    // Set up feed listeners, which update the state whenever 
    // a new status update is received
    useFeed({
        onConnect: () => setState(s => ({ ...s, connected: true, error: false })),
        onMessage: e => console.log('Unrecognized message', e),
        onDisconnect: () => setState(s => ({ ...s, connected: false })),
        onSessionCount: d => setState(s => ({ ...s, sessionCount: d.count })),
        onConnectStatus: handleConnectStatus(u => setState(s => ({ ...s, connectStatus: { ...s.connectStatus, ...u } }))),
        onError: e => setState(s => ({ ...s, error: e }))
    });

    if (!state.connected) {
        // Render disconnection message
        return (<p>Connecting...</p>)
    }

    // Render panel
    return (
        <Container maxWidth="md">
            <Box component="div" className="status-panel">
                <Typography variant="h4" component="h1" gutterBottom>
                    Status Panel
                </Typography>
                <Grid container spacing={3}>
                    <PanelItem title="Active Irma sessions" value={state.sessionCount} />
                    {state.connectStatus && (<>
                        <PanelItem title="Agents online" value={state.connectStatus.agentsOnline} />
                        <PanelItem title="Agents available" value={state.connectStatus.agentsAvailable} />
                        <PanelItem title="Agents on call" value={state.connectStatus.agentsOnCall} />
                        <PanelItem title="Contacts in queue" value={state.connectStatus.contactsInQueue} /></>
                    )}
                </Grid>
            </Box>
        </Container>
    )
};


export default Panel;