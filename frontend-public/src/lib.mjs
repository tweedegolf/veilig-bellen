import "babel-polyfill";
import { render, h } from 'preact';
import App from './components/App';

const veiligBellen = {
    activeElement: null,
    start: async ({ hostname, purpose }) => {
        if (veiligBellen.activeElement !== null) {
            console.error('Element is still active');
            return;
        }

        veiligBellen.activeElement = document.createElement('div');
        veiligBellen.activeElement.setAttribute('class', 'irma-veilig-bellen-overlay');
        document.body.appendChild(veiligBellen.activeElement);

        render(<App
            onClose={() => {
                document.body.removeChild(veiligBellen.activeElement);
                veiligBellen.activeElement = null;
            }}
            hostname={hostname}
            purpose={purpose}
        />, veiligBellen.activeElement);
    },
};

window.veiligBellen = veiligBellen;