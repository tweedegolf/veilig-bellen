import React, { useState } from 'react';

import CssBaseline from '@material-ui/core/CssBaseline';
import Alert from '@material-ui/lab/Alert';

import ContactInfo from './ContactInfo';
import Ccp from './Ccp';

const App = () => {
    const [state, setState] = useState({ mode: 'idle' });
    const [error, setError] = useState(null);

    const onContact = (phonenumber) => {
        setState({ mode: 'connected', phonenumber });
    };

    const onDisclosure = ({ disclosed, purpose }) => {
        setState(state => ({ ...state, mode: 'disclosed', disclosed, purpose }));
    };

    const onConnect = () => {
        // TODO send push message
        setState(state => ({ ...state, mode: 'connected' }));
    };

    const onDisconnect = () => {
        // TODO send push message
        setState({ mode: 'idle' });
    };

    return (
        <CssBaseline>
            <h1>IRMA veilig bellen ({state.mode})</h1>
            {error && <Alert severity="error">{error}</Alert>}
            <ContactInfo {...state} />
            <Ccp {...{ setError, onContact, onDisclosure, onConnect, onDisconnect }} />
        </CssBaseline>
    );
};

export default App;