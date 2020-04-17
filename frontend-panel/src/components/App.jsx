import React from 'react';
import ApiProvider from './ApiProvider';
import Panel from './Panel';
import { useApi } from '../hooks';

const App = () => (
    <ApiProvider>
        <Panel></Panel>
    </ApiProvider>
);


export default App;