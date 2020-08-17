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
            const language = process.env.IRMAJS_LANGUAGE || 'en';
            await handleSession(sessionPtr, {language});
        } catch (e) {
            console.error(e);
            setError();
        }
    };

    return <div className="dialog">
        <button className="button-icon" onClick={onClose}><i class="material-icons">close</i></button>
        <Inner {...{ state, onStartSession, phonenumber: storedPhonenumber }} />
    </div>
};


export default App;