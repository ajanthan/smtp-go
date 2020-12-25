import React,{useState,useEffect} from 'react'
import {Container, Header, Table, List, Portal, Segment, Button, Grid, Label} from 'semantic-ui-react'

function Mails() {
  const [mails,setMails]=useState(new Map())
  const [open,setOpen]=useState(false)
  const [mail,setMailContent]=useState({})
  useEffect(()=>{
    fetch('../mail')
        .then(response => response.json())
        .then(data => {
            let mailMap = new Map();
            data.map( (mail)=> mailMap.set(mail.ID,mail));
            setMails(mailMap);
        })
  },[])
  return (
        <Container style={{ margin: 0 }}>
          <Header as='h2'>Mails</Header>
          <List divided relaxed>
            {[...mails.values()].map((mail)=>
            <List.Item key={mail.ID} onClick={()=> {
                fetch('../mail/'+mail.ID+'/content')
                    .then(response => response.text())
                    .then(data => setMailContent({Subject:mail.Subject,To:mail.To,From:mail.From,Date:mail.Date,Content:data}));
                setOpen(true);
            }}>
                <List.Icon name='mail' size='large' verticalAlign='middle' />
              <List.Content>
                <List.Header as='a'>{mail.Subject}</List.Header>
                <List.Description as='a'>from: {mail.From}</List.Description>
              </List.Content>
            </List.Item>
                  )}
          </List>
            <Portal onClose={()=>setOpen(false)} open={open}>
                <Segment
                    style={{
                        left: '25%',
                        position: 'fixed',
                        top: '8%',
                        zIndex: 1000,
                        overflow: 'auto',
                        maxHeight: '90%',
                    }}
                >
                    <Header>{mail.Subject}</Header>
                    <Table>
                        <Table.Body>
                            <Table.Row>
                                <Table.Cell>
                                <Label>From</Label>
                                </Table.Cell>
                                <Table.Cell>
                                    {mail.From}
                                </Table.Cell>
                            </Table.Row>
                            <Table.Row>
                                <Table.Cell>
                                <Label>To</Label>
                                </Table.Cell>
                                <Table.Cell>{mail.To}</Table.Cell>
                            </Table.Row>
                            <Table.Row>
                                <Table.Cell>
                                <Label>Date</Label>
                                </Table.Cell>
                                <Table.Cell>{mail.Date}</Table.Cell>
                            </Table.Row>
                            <Table.Row>
                                <Table.Cell/>
                                <Table.Cell> <div dangerouslySetInnerHTML={{__html: mail.Content}}/></Table.Cell>
                            </Table.Row>
                        </Table.Body>
                    </Table>
                    <Button
                        content='Close'
                        negative
                        onClick={()=>setOpen(false)}
                    />
                </Segment>
            </Portal>
        </Container>
  );
}

export default Mails;
