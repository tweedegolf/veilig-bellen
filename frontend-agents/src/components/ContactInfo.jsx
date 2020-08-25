import React from 'react';

import Card from '@material-ui/core/Card';
import CardContent from '@material-ui/core/CardContent';
import TableContainer from '@material-ui/core/TableContainer';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableRow from '@material-ui/core/TableRow';
import Link from '@material-ui/core/Link';

const Details = ({ data, urlTemplates }) => (
    <TableContainer>
        <Table>
            <TableBody>
                {data.map(({ key, value }) => (
                    <TableRow key={key}>
                        <TableCell component="th">{key}</TableCell>
                        <TableCell>{
                            urlTemplates && key in urlTemplates 
                                ? <Link href={urlTemplates[key].replace('{}', encodeURIComponent(value))}>{value}</Link>
                                : value
                        }</TableCell>
                    </TableRow>
                ))}
            </TableBody>
        </Table>
    </TableContainer>
)

const ContactInfo = ({ phonenumber, disclosed, purpose, urlTemplates }) => (
    phonenumber ? <Card>
        <CardContent>
            <h2>Attributen van contactpersoon</h2>
            <Details urlTemplates={urlTemplates} data={
                [{ key: "Telefoonnummer", value: phonenumber }, { key: "Doel", value: purpose }].concat(
                    disclosed ? disclosed.flat().map(attr => ({ key: attr.id, value: attr.rawvalue })) : []
                )
            } />
        </CardContent>
    </Card> : null
);

export default ContactInfo;