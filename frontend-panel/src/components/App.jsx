import React from 'react';
import ApiProvider from './ApiProvider';
import Panel from './panel/Panel';

const App = () => (
    <ApiProvider>
        <Panel></Panel>
    </ApiProvider>
);


export default App;