import React from 'react';
import {Container, Icon,Menu} from 'semantic-ui-react';


export default () =>{
   return( <Menu>
        <Container>
            <Menu.Item as="a" header>
                <Icon.Group size='large'>
                    <Icon size='big' name='circle outline' />
                    <Icon name='envelope open' />
                </Icon.Group>
            </Menu.Item>

            <Menu.Menu position="right">
                <Menu.Item as="a" name="login">
                    Login
                </Menu.Item>

                <Menu.Item as="a" name="register">
                    Register
                </Menu.Item>
            </Menu.Menu>
        </Container>
    </Menu>)
}