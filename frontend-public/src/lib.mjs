import "babel-polyfill";
import axios from 'axios';
import qrcode from 'qrcode-terminal';
import { handleSession } from '@privacybydesign/irmajs';

const veiligBellen = {
    start: async ({ url, purpose }) => {
        // TODO better error handling
        const response = await axios.get(`${url}/session`, { params: { purpose } });
        
        if (response.status !== 200) {
            console.error(response.statusCode);
            return;
        }

        const { sessionPtr, phonenumber } = response.data;

        await handleSession(sessionPtr);

        console.log(`Please place a call now to: ${phonenumber}`);
        qrcode.generate(`tel:${phonenumber}`);
    },
};

window.veiligBellen = veiligBellen;