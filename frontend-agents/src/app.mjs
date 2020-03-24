import React from 'react';
import ReactDOM from 'react-dom';

import App from './components/App';

window.addEventListener('load', () => {
    const container = window.document.getElementById('container');

    ReactDOM.render(<App />, container);
});