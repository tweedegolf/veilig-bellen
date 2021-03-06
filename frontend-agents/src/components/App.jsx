import React, { useState } from 'react';
import axios from 'axios';

import CssBaseline from '@material-ui/core/CssBaseline';
import Alert from '@material-ui/lab/Alert';
import Grid from '@material-ui/core/Grid';
import Button from '@material-ui/core/Button';

import Box from '@material-ui/core/Box';

import ContactInfo from './ContactInfo';
import Ccp from './Ccp';


const updateStatus = async (backendUrl, secret, status) => {
    if (!secret) {
        return;
    }

    const res = await axios.post(`${backendUrl}/session/update`, { secret, status });

    if (res.status !== 200) {
        throw new Error("Failed to update status");
    }
};

const destroySession = async (backendUrl, secret) => {
    if (!secret) {
        return;
    }

    const res = await axios.post(`${backendUrl}/session/destroy`, { secret });

    if (res.status !== 200) {
        throw new Error("Failed to destroy session");
    }
}

const getDisclosure = async (backendUrl, secret) => {
    const response = await axios.post(`${backendUrl}/disclose`, { secret });
    if (response.status === 200) {
        return response.data;
    } else {
        throw new Error('Failed to disclose');
    }
};

const App = ({ backendUrl, ccpHost, urlTemplates, metricsUrl }) => {
    const [state, setState] = useState({ mode: 'unauthorized' });
    const [error, setErrorBase] = useState(null);

    const setError = (msg) => {
        setTimeout(() => {
            setErrorBase((originalError) => {
                if (msg === originalError) {
                    return null; // Reset if no other error occurred.
                } else {
                    return originalError; // Do not reset if error has change in the meantime.
                }
            })
        }, 10000);

        setErrorBase(msg);
    };

    const resetState = (state) => {
        setError(null);
        setState(state);
    };

    const onAgent = () => {
        // When logged in and agent information is available, set to idle.
        resetState({ mode: 'idle' });
    };

    const onContact = async (secret, phonenumber) => {
        // When contact is standing by, establish a connection and retrieve the disclosed facts.
        resetState({ mode: 'establishing', secret, phonenumber });

        try {
            const { disclosed, purpose } = await getDisclosure(backendUrl, secret);
            setState(state => ({ ...state, mode: 'disclosed', disclosed, purpose }));
        } catch (e) {
            console.error(e);
            setError("Failed to retrieve disclosed attributes.");
        }
    };

    const onConnect = () => {
        // When contact has connected, show thusly.
        setState(state => {
            updateStatus(backendUrl, state.secret, 'CONNECTED');
            return { ...state, mode: 'connected' };
        });
    };

    const onDisconnect = () => {
        // When contact has disconnected, go back to idle.
        setState(state => {
            updateStatus(backendUrl, state.secret, 'DONE');
            return { ...state, mode: 'disconnected' };
        });
    };

    const onDestroy = () => {
        setState(state => {
            try {
                destroySession(backendUrl, state.secret);
            } catch (e) {
                console.error(e);
            }
            return { mode: 'idle' };
        })
    };

    return (
        <CssBaseline>
            <h1>IRMA veilig bellen ({state.mode})</h1>
            {error && <Alert severity="error">{error}</Alert>}
            {state.mode === 'unauthorized' && (<Alert severity="warning">You are yet unauthorized and are required to log in using the pop-up.</Alert>)}
            <Grid container spacing={2}>
                <Grid className="contactinfo" item xs={6}>
                    <Ccp {...{ setError, onAgent, onContact, onConnect, onDisconnect, onDestroy, ccpHost }} />
                </Grid>
                <Grid className="contactinfo" item xs={6}>
                    <ContactInfo {...state} urlTemplates={urlTemplates} />
                </Grid>
            </Grid>
            {metricsUrl &&
                <Box
                    zIndex="tooltip"
                    position="absolute"
                    right="1rem"
                    top="1rem" >
                    <Button
                        color="primary"
                        variant="outlined"
                        target="_blank"
                        noopener
                        noreferrer
                        href={metricsUrl}>
                        Metrics
                    </Button>
                </Box>}
        </CssBaseline>
    );
};

export default App;
