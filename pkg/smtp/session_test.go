package smtp

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github/ajanthan/smtp-go/pkg/storage"
	"net"
	"net/smtp"
	"testing"
)

func TestSession_HandleReset(t *testing.T) {
	mailChan := make(chan storage.Envelope)
	address := "localhost:20246"
	go func() {
		ln, err := net.Listen("tcp", address)
		assert.NoError(t, err)
		conn, err := ln.Accept()
		assert.NoError(t, err)
		session := &Session{
			Conn:   conn,
			Server: "localhost",
		}
		err = session.Start()
		if err != nil {
			session.HandleUnknownError(err)
			assert.Fail(t, err.Error())
			close(mailChan)
		}

		mail, err := session.GetMail()
		if err != nil {
			session.HandleUnknownError(err)
			assert.Fail(t, err.Error())
			close(mailChan)
		}
		mailChan <- mail
	}()

	c, err := smtp.Dial(address)
	assert.NoError(t, err)
	err = c.Mail("test0@test.com")
	assert.NoError(t, err)
	err = c.Rcpt("rtest0@test.com")
	assert.NoError(t, err)
	err = c.Reset()
	assert.NoError(t, err)
	err = c.Mail("test1@test.com")
	assert.NoError(t, err)
	err = c.Rcpt("rtest1@test.com")
	assert.NoError(t, err)
	wc, err := c.Data()
	assert.NoError(t, err)
	_, err = fmt.Fprint(wc, "Hi")
	assert.NoError(t, err)
	err = wc.Close()
	assert.NoError(t, err)
	err = c.Quit()
	assert.NoError(t, err)

	mail := <-mailChan
	assert.Equal(t, "test1@test.com", mail.Sender)
	assert.Equal(t, "rtest1@test.com", mail.Recipient[0])
}
