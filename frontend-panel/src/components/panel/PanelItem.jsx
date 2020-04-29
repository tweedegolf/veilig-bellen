import React from 'react';
import { Grid, Card, CardContent, Container, Typography } from '@material-ui/core';

const PanelItem = ({ title, value }) => {
    return (
        <Grid item xs={6}>
            <Card>
                <CardContent>
                    <Container>
                        <Typography variant="h5" component="h1">
                            {title}
                        </Typography>
                        <Typography variant="h1" component="p">
                            {value}
                        </Typography>
                    </Container>
                </CardContent>
            </Card>
        </Grid>);
};

export default PanelItem;