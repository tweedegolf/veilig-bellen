import "babel-polyfill";
import axios from 'axios';
import qrcode from 'qrcode-terminal';
import { handleSession } from '@privacybydesign/irmajs';

const veiligBellen = {
    start: async ({ hostname, purpose }) => {
        // TODO better error handling
        const response = await axios.get(`https://${hostname}/session`, { params: { purpose } });

        if (response.status !== 200) {
            console.error(response.statusCode);
            return;
        }

        const { sessionPtr, phonenumber, dtmf } = response.data;

        const client = new WebSocket(`wss://${hostname}/session/status?dtmf=${encodeURIComponent(dtmf)}`);

        client.addEventListener('error', (error) => {
            console.log('Connect Error: ', error);
        });

        client.addEventListener('open', () => {
            console.log('Connection established');
        });

        client.addEventListener('message', (event) => {
            console.log('Message', event);
        });

        await handleSession(sessionPtr);

        console.log(`Please place a call now to: ${phonenumber}`);
        qrcode.generate(`tel:${phonenumber}`);
    },
};

window.veiligBellen = veiligBellen;