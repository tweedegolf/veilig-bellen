import React, { useCallback } from 'react';

import 'amazon-connect-streams';

const Ccp = ({ setError, onContact, onAgent, onConnect, onDisconnect, onDestroy, ccpHost }) => {
    const ccpUrl = `https://${ccpHost}/connect/ccp-v2`;

    const containerRef = useCallback(element => {
        if (element !== null) {
            // Clear the localStorage to ensure that the popup is shown.
            if (window.localStorage !== null) {
                window.localStorage.removeItem('connectPopupManager::connect::loginPopup');
            }

            connect.core.initCCP(element, {
                ccpUrl,
                loginPopup: true,
                loginPopupAutoClose: true,
                softphone: {
                    allowFramedSoftphone: true,
                }
            });

            connect.core.eventBus.subscribe("ack_timeout", () => {
                setError("Failed to authenticate, please log in using the popup.");
            });

            connect.agent((_agent) => void onAgent());

            connect.contact(async (contact) => {
                console.log('contact', contact);

                const callAttributes = contact.getAttributes();

                contact.onConnected(() => void onConnect());
                contact.onEnded(() => void onDisconnect());
                contact.onDestroy(() => void onDestroy());

                onContact(callAttributes.session_secret.value, callAttributes.phonenumber.value);
            });
        }
    }, []);

    return (<div className="ccp" ref={containerRef} ></div>);
};

export default Ccp;