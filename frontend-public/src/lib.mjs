import axios from 'axios';
import { handleSession } from '@privacybydesign/irmajs';

const veiligBellen = {
    start: async ({ url, purpose }) => {
        // TODO better error handling
        const response = await axios.get(`${url}/session`, { params: { purpose } });

        if (response.status !== 200) {
            console.error(response.statusCode);
            return;
        }

        console.log(await handleSession(response.data));
    },
};

window.veiligBellen = veiligBellen;