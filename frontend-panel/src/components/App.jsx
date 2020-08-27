import React from 'react';
import ApiProvider from './ApiProvider';
import Panel from './panel/Panel';

const App = ({backendFeedUrl}) => (
    <ApiProvider backendFeedUrl={backendFeedUrl}>
        <Panel></Panel>
    </ApiProvider>
);


export default App;