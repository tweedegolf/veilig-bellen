import React, { useState } from 'react';

import State from './State';
import Ccp from './Ccp';

const App = () => {
    const [state, setState] = useState('idle');
    const [disclosure, setDisclosure] = useState(null);

    const onContact = (/*number, purpose*/) => {
        setState('connected');
    };

    const onDisclosure = (d) => {
        setState('disclosed');
        setDisclosure(d);
    };

    const onDisconnect = () => {
        setState('idle');
        setDisclosure(null);
    };

    return (
        <>
            <h1>IRMA veilig bellen</h1>
            <State {...{ state, disclosure }} />
            <Ccp {...{ onContact, onDisclosure, onDisconnect }} />
        </>
    );
};

export default App;