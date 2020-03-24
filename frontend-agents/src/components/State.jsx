import React from 'react';

import Disclosure from './Disclosure';

const State = ({ state, disclosure }) => (
    <table>
        <tbody>
            <tr><th>State</th><td>{state}</td></tr>
            <tr><th>Disclosure</th><td><Disclosure disclosure={disclosure} /></td></tr>
        </tbody>
    </table>
);

export default State;