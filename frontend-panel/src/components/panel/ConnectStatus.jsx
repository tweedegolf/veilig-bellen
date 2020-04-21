import React from 'react';
import PanelItem from './PanelItem';
import { Grid, Paper } from '@material-ui/core';

// Render Amazon connect status updates
const ConnectStatus = ({ status: s }) => (
    <>
        {/* <PanelItem title="Last Connect update" value={s.lastUpdate ? s.lastUpdate.toLocaleDateString() : '-'} /> */}
        <PanelItem title="Agents online" value={s.agentsOnline} />
        <PanelItem title="Agents available" value={s.agentsAvailable} />
        <PanelItem title="Agents on call" value={s.agentsOnCall} />
        <PanelItem title="Contacts in queue" value={s.contactsInQueue} />
    </>);


export default ConnectStatus;