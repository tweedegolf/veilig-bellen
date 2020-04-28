import React, { useCallback } from 'react';
import axios from 'axios';

import 'amazon-connect-streams';

const Ccp = ({ setError, onContact, onDisclosure, onConnect, onDisconnect, backendUrl, ccpHost }) => {
    const backendDiscloseUrl = `${backendUrl}/disclose`;
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
                console.log('auth failure');
            });

            connect.core.onSoftphoneSessionInit(function ({ connectionId }) {
                var softphoneManager = connect.core.getSoftphoneManager();
                if (softphoneManager) {
                    var session = softphoneManager.getSession(connectionId);
                    // You can use this rtc session for stats analysis 

                    console.log('session', session);
                }
            });

            connect.agent((agent) => {
                console.log('agent', agent);
                console.log("agent-conf", agent.getConfiguration());
            });

            connect.contact(async (contact) => {
                console.log('contact', contact);

                contact.onConnected(() => {
                    onConnect();
                });

                contact.onEnded(() => {
                    onDisconnect();
                });

                const callAttributes = contact.getAttributes();

                onContact(callAttributes.phonenumber.value);

                console.log('callAttributes', callAttributes);
                const response = await axios.get(vbServerDisclose, {
                    params: {
                        secret: callAttributes.session_secret.value,
                    },
                });

                if (response.status === 200) {
                    onDisclosure(response.data);
                } else {
                    setError('Failed to retrieve disclosed data');
                }
            });
        }
    }, []);

    return (<div className="ccp" ref={containerRef} ></div>);
};

export default Ccp;