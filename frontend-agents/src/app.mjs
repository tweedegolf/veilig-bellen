import 'amazon-connect-streams';
import axios from 'axios';

const vbServerDisclose = 'https://backend.veiligbellen.test.tweede.golf/disclose';

window.addEventListener('load', () => {
    const containerDiv = window.document.getElementById('ccp');
    const ccpUrl = 'https://sarif.awsapps.com/connect/ccp-v2';

    connect.core.initCCP(containerDiv, {
        ccpUrl,
        loginPopup: false,
        softphone: {
            allowFramedSoftphone: true,
        }
    });

    connect.core.onAuthFail(() => {
        console.log('auth failure');
    });

    connect.core.onSoftphoneSessionInit(function ({ connectionId }) {
        var softphoneManager = connect.core.getSoftphoneManager();
        if (softphoneManager) {
            // access session
            var session = softphoneManager.getSession(connectionId);
            // YOu can use this rtc session for stats analysis 

            console.log('session', session);
        }
    });

    connect.agent((agent) => {
        console.log('agent', agent);
        console.log("agent-conf", agent.getConfiguration());
    });

    connect.contact(async (contact) => {
        console.log('contact', contact);

        const attributes = contact.getAttributes();

        console.log('attributes', attributes);
        const response = await axios.get(vbServerDisclose, {
            params: {
                secret: attributes.session_secret.value,
            },
        });

        if (response.status === 200) {
            console.log(response.data);
        }
    });
});
