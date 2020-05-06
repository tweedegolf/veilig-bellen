import React, { useCallback } from 'react';

import 'amazon-connect-streams';

const Ccp = ({ setError, onContact, onConnect, onDisconnect, ccpHost }) => {
    const ccpUrl = `https://${ccpHost}/connect/ccp-v2`;

    const containerRef = useCallback(element => {
        if (element !== null) {
            connect.core.initCCP(element, {
                ccpUrl,
                loginPopup: false,
                softphone: {
                    allowFramedSoftphone: true,
                }
            });

            connect.core.onAuthFail(() => {
                setError('auth_failure');
            });

            connect.agent((agent) => {
                console.log('agent', agent);
                console.log('agent-conf', agent.getConfiguration());
            });

            connect.contact(async (contact) => {
                console.log('contact', contact);

                const callAttributes = contact.getAttributes();

                contact.onConnected(() => { onConnect(); });
                contact.onEnded(() => { onDisconnect(); });

                onContact(callAttributes.session_secret.value, callAttributes.phonenumber.value);
            });
        }
    }, []);

    return (<div className="ccp" ref={containerRef} ></div>);
};

export default Ccp;