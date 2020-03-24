import React, { useCallback } from 'react';
import axios from 'axios';

import 'amazon-connect-streams';

const vbServerDisclose = 'https://backend.veiligbellen.test.tweede.golf/disclose';
const ccpUrl = 'https://sarif.awsapps.com/connect/ccp-v2';

const Ccp = ({ setError, onContact, onDisclosure, onConnect, onDisconnect }) => {
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

                const attributes = contact.getAttributes();

                onContact(attributes.phonenumber.value);

                console.log('attributes', attributes);
                const response = await axios.get(vbServerDisclose, {
                    params: {
                        secret: attributes.session_secret.value,
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