import { h } from 'preact';
import { useState, useCallback, useEffect } from 'preact/hooks';

import axios from 'axios';
import { handleSession, detectUserAgent } from '@privacybydesign/irmajs';

import Inner from './Inner';

// States for which we should not render the popup
const hiddenStates = [
    'INIT', 'IRMA-INITIALIZED', 
]

const App = ({ hostname, purpose, onClose, irmaJsLang }) => {
    const [state, setState] = useState('INIT');
    const [closing, setClosing] = useState(false);

    const getUserAgent = useCallback(detectUserAgent, []);
    const isHidden = useCallback(() => hiddenStates.includes(state), [state]);

    const setError = () => { setState('ERROR'); };

    const doClose = () => {
        if (closing) {
            return;
        }
        setClosing(true);
        onClose();
    }

    const onStartSession = async () => {
        try {
            const response = await axios.get(`https://${hostname}/session`, { params: { purpose } });

            if (response.status !== 200) {
                setError();
                return;
            }

            const { sessionPtr, statusToken } = response.data;

            const client = new WebSocket(`wss://${hostname}/session/status?statusToken=${encodeURIComponent(statusToken)}`);

            client.addEventListener('error', (error) => {
                console.error('Connect Error: ', error);
                setError();
            });

            client.addEventListener('open', () => {
                console.log('Connection established');
            });

            client.addEventListener('message', (event) => {
                if (event.data !== '') {
                    setState(event.data);
                    console.log('Message', event.data);
                }
            });


            const language = irmaJsLang || 'en';
            await handleSession(sessionPtr, { language });
        } catch (e) {
            if (e === 'CANCELLED') {
                setState(e)
            } else {
                console.error(e);
                setError();
            }
        }
    };

    // Start IRMA session immediately
    useEffect(onStartSession, []);

    // Close popup on cancellation
    if (state === 'IRMA-CANCELLED' || state === 'CANCELLED') {
        doClose();
        return null;
    }

    // Close popup on connection if on mobile device
    if (state === 'IRMA-CONNECTED' && getUserAgent() !== 'Desktop') {
        doClose();
        return null;
    }

    if (isHidden()) {
        return null;
    }

    return <div className="irma-veilig-bellen-overlay"><div className="dialog">
        <button className="button-icon" onClick={onClose}><i class="material-icons">close</i></button>
        <Inner {...{ state }} />
    </div></div>
};


export default App;
