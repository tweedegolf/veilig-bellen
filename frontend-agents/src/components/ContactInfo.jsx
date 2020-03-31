import React from 'react';

import Card from '@material-ui/core/Card';
import CardContent from '@material-ui/core/CardContent';
import TableContainer from '@material-ui/core/TableContainer';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableRow from '@material-ui/core/TableRow';

const Details = ({ data }) => (
    <TableContainer>
        <Table>
            <TableBody>
                {data.map(({ key, value }) => (
                    <TableRow key={key}>
                        <TableCell component="th">{key}</TableCell>
                        <TableCell>{value}</TableCell>
                    </TableRow>
                ))}
            </TableBody>
        </Table>
    </TableContainer>
)

const ContactInfo = ({ phonenumber, disclosed, purpose }) => (
    phonenumber ? <Card>
        <CardContent>
            <Details data={
                [{ key: "phonenumber", value: phonenumber }, { key: "purpose", value: purpose }].concat(
                    disclosed ? disclosed.flat().map(attr => ({ key: attr.id, value: attr.rawvalue })) : []
                )
            } />
        </CardContent>
    </Card> : null
);

export default ContactInfo;