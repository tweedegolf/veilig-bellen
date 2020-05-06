import React, { useState } from 'react';
import axios from 'axios';

import CssBaseline from '@material-ui/core/CssBaseline';
import Alert from '@material-ui/lab/Alert';

import ContactInfo from './ContactInfo';
import Ccp from './Ccp';

const updateStatus = async (backendUrl, secret, status) => {
    if (!secret) {
        return;
    }

    const res = await axios.get(`${backendUrl}/session/update`, {
        params: {
            secret,
            status,
        },
    });

    if (res.status !== 200) {
        throw new Error("Failed to update status");
    }
};

const getDisclosure = async (backendUrl, secret) => {
    const response = await axios.get(`${backendUrl}/disclose`, { params: { secret } });
    if (response.status === 200) {
        return response.data;
    } else {
        throw new Error('Failed to disclose');
    }
};

const App = ({ backendUrl, ccpHost }) => {
    const [state, setState] = useState({ mode: 'idle' });
    const [error, setError] = useState(null);

    const resetState = (state) => {
        setError(null);
        setState(state);
    };

    const onContact = async (secret, phonenumber) => {
        resetState({ mode: 'establishing', secret, phonenumber });

        try {
            const { disclosed, purpose } = await getDisclosure(backendUrl, secret);
            setState(state => ({ ...state, mode: 'disclosed', disclosed, purpose }));
        } catch (e) {
            console.error(e);
            setError("Failed to retrieve disclosed attributes");
        }
    };

    const onConnect = () => {
        setState(state => {
            updateStatus(backendUrl, state.secret, 'CONNECTED');
            return { ...state, mode: 'connected' };
        });
    };

    const onDisconnect = () => {
        setState(state => {
            updateStatus(backendUrl, state.secret, 'DONE');
            return { mode: 'idle' };
        });
    };

    return (
        <CssBaseline>
            <h1>IRMA veilig bellen ({state.mode})</h1>
            {error && <Alert severity="error">{error}</Alert>}
            <ContactInfo {...state} />
            <Ccp {...{ setError, onContact, onConnect, onDisconnect, ccpHost }} />
        </CssBaseline>
    );
};

export default App;