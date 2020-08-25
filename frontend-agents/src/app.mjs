import "regenerator-runtime/runtime.js";
import React from 'react';
import ReactDOM from 'react-dom';

import App from './components/App';

const tryParseJSON = (str) => {
    try {
        return JSON.parse(str);
    } catch (_) {
        return false;
    }
}

window.addEventListener('load', () => {
    const container = window.document.getElementById('container');


    ReactDOM.render(<App
        backendUrl={process.env.BACKEND_URL}
        ccpHost={process.env.CCP_HOST}
        urlTemplates={tryParseJSON(process.env.URL_TEMPLATES)}
        metricsUrl={process.env.METRICS_URL}
    />, container);
});
