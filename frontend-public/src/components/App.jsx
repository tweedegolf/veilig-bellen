import { h } from 'preact';
import { useState } from 'preact/hooks';

import axios from 'axios';
import { handleSession } from '@privacybydesign/irmajs';

import Inner from './Inner';

const App = ({ hostname, purpose, onClose }) => {
    const [state, setState] = useState('INIT');
    const [storedPhonenumber, setPhonenumber] = useState(null);

    const setError = () => { setState('ERROR'); };

    const onStartSession = async () => {
        const response = await axios.get(`https://${hostname}/session`, { params: { purpose } });

        if (response.status !== 200) {
            setError();
            return;
        }

        const { sessionPtr, phonenumber, dtmf } = response.data;
        setPhonenumber(phonenumber);

        const client = new WebSocket(`wss://${hostname}/session/status?dtmf=${encodeURIComponent(dtmf)}`);

        client.addEventListener('error', (error) => {
            console.error('Connect Error: ', error);
            setError();
        });

        client.addEventListener('open', () => {
            console.log('Connection established');
        });

        client.addEventListener('message', (event) => {
            setState(event.data);
            console.log('Message', event.data);
        });

        try {
            await handleSession(sessionPtr);
        } catch (e) {
            console.error(e);
            setError();
        }
    };

    return <div className="dialog">
        <button onClick={onClose}>Close</button>
        <Inner {...{ state, onStartSession, phonenumber: storedPhonenumber }} />
    </div>
};


export default App;