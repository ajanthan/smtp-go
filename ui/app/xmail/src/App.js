import React, { Fragment } from 'react';
import { Grid, Icon, Menu} from 'semantic-ui-react';

import Mails from "./Mails";
import Header from "./Header";

const App = () => (
    <Fragment>
        <Header />
        <Grid container columns={2}>
            <Grid.Row>
                <Grid.Column width={4}>
        <Menu vertical>
            <Menu.Item>
                <Input placeholder='Search...' />
            </Menu.Item>
            <Menu.Item
                name='Mail'>
                Mail
            </Menu.Item>
        </Menu>
                </Grid.Column>
                <Grid.Column width={12}>
            <Mails />
                </Grid.Column>
            </Grid.Row>
        </Grid>
    </Fragment>
);

export default App;