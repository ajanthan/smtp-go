package smtp

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/stretchr/testify/assert"
	"github/ajanthan/smtp-go/pkg/storage"
	"io/ioutil"
	"net"
	"net/smtp"
	"net/textproto"
	"strings"
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
			Conn:   textproto.NewConn(conn),
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

	err = sendMessageBody(wc, "test1@test.com", "rtest1@test.com", "Test", strings.NewReader("Hi"))
	assert.NoError(t, err)

	err = c.Quit()
	assert.NoError(t, err)

	mail := <-mailChan
	assert.Equal(t, "test1@test.com", mail.Sender)
	assert.Equal(t, "rtest1@test.com", mail.Recipient[0])
}

func TestSession_HandleStartTLS(t *testing.T) {
	cert, err := tls.LoadX509KeyPair("../../resources/localhost.crt", "../../resources/localhost.pkcs8")
	assert.NoError(t, err)

	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}
	caPool := x509.NewCertPool()
	pem, err := ioutil.ReadFile("../../resources/localhost.crt")
	assert.NoError(t, err)

	caPool.AppendCertsFromPEM(pem)

	clientTLSConfig := tlsConfig.Clone()
	clientTLSConfig.ServerName = "localhost"
	clientTLSConfig.RootCAs = caPool

	mailChan := make(chan storage.Envelope)
	address := "localhost:20246"
	go func() {

		ln, err := net.Listen("tcp", address)
		assert.NoError(t, err)
		conn, err := ln.Accept()
		assert.NoError(t, err)
		session := &Session{
			conn:      &conn,
			Conn:      textproto.NewConn(conn),
			Server:    "localhost",
			TLSConfig: tlsConfig,
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

	err = c.Hello("localhost")
	assert.NoError(t, err)

	isTLSSupported, _ := c.Extension("STARTTLS")
	assert.True(t, isTLSSupported)

	err = c.StartTLS(clientTLSConfig)
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

	err = sendMessageBody(wc, "test1@test.com", "rtest1@test.com", "Test", strings.NewReader("Hi"))
	assert.NoError(t, err)

	err = c.Quit()
	assert.NoError(t, err)

	mail := <-mailChan
	assert.Equal(t, "test1@test.com", mail.Sender)
	assert.Equal(t, "rtest1@test.com", mail.Recipient[0])
}
