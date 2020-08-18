import { Fragment, h } from 'preact';

const Inner = ({ state }) => {
    switch (state) {
        case 'IRMA-INITIALIZED':
        case 'IRMA-CONNECTED':
            return <p>Volg de instructie in de IRMA interactie.</p>;
        case 'IRMA-DONE':
            return <Fragment>
                <h2>U kunt ons nu beveiligd bellen met uw mobiele telefoon</h2>
                <p>
                    We hebben uw gegevens correct ontvangen. U kunt nu met ons bellen via de IRMA-app.<br />
                    Nadat u op 'bellen' heeft gedrukt hoort u eerst enkele tonen. Daarna bent u verbonden en staat u in de wachtrij. <br />
                </p>
                <p>We nemen zo spoedig mogelijk op.</p>

            </Fragment>;
        case 'IRMA-CANCELLED':
        case 'CANCELLED':
            return <p>U hebt de IRMA interactie gestopt.</p>;
        case 'CALLED':
            return <p>U bent succesvol met ons verbonden. We helpen u zo spoedig mogelijk.</p>;
        case 'CONNECTED':
            return <p>U bent nu in gesprek met de medewerker.</p>;
        case 'DONE':
            return <p>Uw gesprek is voltooid.</p>;
        case 'IRMA-UNREACHABLE':
            return <p>Uw sessie is niet bereikbaar. Mogelijk is deze verlopen.</p>;
        default:
        case 'ERROR':
            return <p>Er ging iets mis, probeer het opnieuw.</p>;
    }
};


export default Inner;