import React from 'react';
import ReactDOM from 'react-dom';

import App from './components/App';

window.addEventListener('load', () => {
    const container = window.document.getElementById('container');

    ReactDOM.render(<App
        backendUrl={process.env.BACKEND_URL}
        ccpHost={process.env.CCP_HOST}
    />, container);
});