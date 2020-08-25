import React from 'react';
import ReactDOM from 'react-dom';

import App from './components/App';

window.addEventListener('load', () => {
    const container = window.document.getElementById('container');
    const backendFeedUrl = process.env.BACKEND_FEED_URL;
    ReactDOM.render(<App backendFeedUrl={backendFeedUrl}/>, container);
});