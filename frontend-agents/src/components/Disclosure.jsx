import React from 'react';

const Disclosure = ({ disclosure }) => {
    if (disclosure === null) {
        return null;
    } else {
        return (
            <table>
                <tbody>
                    {disclosure.flat().map(attr => (
                        <tr>
                            <th>{attr.id}</th>
                            <td>{attr.value.nl || attr.rawvalue}</td>
                        </tr>
                    ))}
                </tbody>
            </table>
        )
    }
};

export default Disclosure;