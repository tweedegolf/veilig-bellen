import { h } from 'preact';
import { useState, useCallback, useEffect } from 'preact/hooks';

import axios from 'axios';
import { handleSession, detectUserAgent } from '@privacybydesign/irmajs';

import Inner from './Inner';
import { initFeed } from '../feed';

// States for which we should not render the popup
const hiddenStates = [
    'INIT', 'IRMA-INITIALIZED',
]

const App = ({ hostname, purpose, onClose, irmaJsLang }) => {
    const [state, setState] = useState('INIT');
    const [closing, setClosing] = useState(false);
    const [feed, setFeed] = useState(null);

    const getUserAgent = useCallback(detectUserAgent, []);
    const isHidden = useCallback(() => hiddenStates.includes(state), [state]);

    const setError = () => { setState('ERROR'); };

    const doClose = () => {
        if (closing) {
            return;
        }
        setClosing(true);
        if(feed) {
            feed.closeFeed();
        }
        onClose();
    }

    const setupFeed = (statusToken) => {
        if(feed) {
            feed.closeFeed();
        }

        const feedListener = {
            onError: (error) => {
                console.error('Connect Error: ', error);
                setError();
            },
            onMessage: (event) => {
                if (event.data !== '') {
                    setState(event.data);
                    console.log('Message', event.data);
                }
            },
            onConnect: () => console.log('Connection Established')
        };
        const feed = initFeed(`wss://${hostname}/session/status?statusToken=${encodeURIComponent(statusToken)}`);
        setFeed(feed);
        feed.registerFeedListener(feedListener);
    }

    const onStartSession = async () => {
        try {
            const response = await axios.get(`https://${hostname}/session`, { params: { purpose } });

            if (response.status !== 200) {
                setError();
                return;
            }

            const { sessionPtr, statusToken } = response.data;
            setupFeed(statusToken);

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
