import { Fragment, h } from 'preact';
import { useEffect, useState, useCallback } from 'preact/hooks';
import QRCode from 'qrcode';

const makePhoneLink = (phonenumber) => `tel:${phonenumber}`;

const Inner = ({ state, onStartSession, phonenumber }) => {
    const [qrcodeSvg, setQrcodeSvg] = useState(null);

    useEffect(() => {
        if (phonenumber === null) {
            return;
        }

        QRCode.toString(makePhoneLink(phonenumber), { format: 'svg' }).then((str, err) => {
            if (!err) {
                setQrcodeSvg(str);
            }
        });
    }, [phonenumber]);

    const qrcodeContainer = useCallback(node => {
        if (node !== null) {
            node.innerHTML = qrcodeSvg;
        }
    }, [qrcodeSvg]);

    switch (state) {
        case 'INIT':
            return <Fragment>
                <p>Tekst over IRMA en knop</p>
                <button onClick={onStartSession}>Bel!</button>
            </Fragment>;
        case 'IRMA-INITIALIZED':
        case 'IRMA-CONNECTED':
            return <p>Volg de instructie in de IRMA interactie.</p>;
        case 'IRMA-DONE':
            return <Fragment>
                <p>Start nu het telefoongesprek met uw bel-applicatie.</p>
                <p>
                    Indien u geen telefoon hebt gebruikt om uw IRMA sessie door te zetten,
                    kunt u <a href={makePhoneLink(phonenumber)}>{phonenumber}</a> bellen of de volgende QR-code inscannen met een applicatie:
                </p>
                <div className="phonenumber-qrcode" ref={qrcodeContainer} >laden...</div>
                <p>U hoort eerst een aantal tonen, waarna u in de wachtrij geplaatst wordt.</p>
            </Fragment>;
        case 'IRMA-CANCELLED':
            return <p>U hebt de IRMA interactie gestopt.</p>;
        case 'CALLED':
            return <p>
                U bent succesvol verbonden met de wachtrij.
                Een medewerker neemt zo spoedig mogelijk op.
            </p>;
        case 'CONNECTED':
            return <p>U bent nu in gesprek met de medewerker.</p>;
        case 'DONE':
            return <p>Uw gesprek is voltooid.</p>;
        default:
        case 'ERROR':
            return <p>Er ging iets mis, probeer het opnieuw.</p>;
    }
};


export default Inner;