package smtp

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/tls"
	"crypto/x509"
	"fmt"
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
	mailChan := make(chan *storage.Envelope)
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
	serverTLSConfig, clientTLSConfig, err := getTestTLSConfig()
	assert.NoError(t, err)

	mailChan := make(chan *storage.Envelope)
	address := "localhost:20247"
	go startTesTLStServer(t, address, serverTLSConfig, mailChan, []string{}, nil)

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

type TestAuthService struct {
	userDB map[string][]byte
}

func NewTestAuthService() *TestAuthService {
	return &TestAuthService{
		make(map[string][]byte),
	}
}
func (auth *TestAuthService) Authenticate(username string, password []byte) error {
	secret, ok := auth.userDB[username]
	if !ok || len(secret) != len(password) {
		return NewInvalidCredentialError("invalid credential")
	}
	for i, b := range password {
		if b != secret[i] {
			return NewInvalidCredentialError("invalid credential")
		}
	}
	return nil
}
func (auth *TestAuthService) ValidateHMAC(username string, msg []byte, code []byte) error {
	secret, ok := auth.userDB[username]
	if !ok {
		return NewInvalidCredentialError("invalid credential")
	}
	hmacEncoder := hmac.New(md5.New, secret)
	hmacEncoder.Write(msg)
	expectedCode := make([]byte, 0, hmacEncoder.Size())
	expectedCode = hmacEncoder.Sum(expectedCode)
	expectedCodeStr := []byte(fmt.Sprintf("%x", expectedCode))

	if len(expectedCodeStr) != len(code) {
		return NewInvalidCredentialError("invalid credential")
	}
	for i, b := range code {
		if b != expectedCodeStr[i] {
			return NewInvalidCredentialError("invalid credential")
		}
	}
	return nil
}
func (auth *TestAuthService) AddUser(username string, password []byte) error {
	_, ok := auth.userDB[username]
	if ok {
		return fmt.Errorf("user %s already exisits", username)
	}
	auth.userDB[username] = password
	return nil
}
func getTestTLSConfig() (*tls.Config, *tls.Config, error) {
	cert, err := tls.LoadX509KeyPair("../../resources/localhost.crt", "../../resources/localhost.pkcs8")
	if err != nil {
		return nil, nil, err
	}
	serverTLSConfig := &tls.Config{Certificates: []tls.Certificate{cert}}

	caPool := x509.NewCertPool()
	pem, err := ioutil.ReadFile("../../resources/localhost.crt")
	if err != nil {
		return nil, nil, err
	}
	caPool.AppendCertsFromPEM(pem)
	clientTLSConfig := serverTLSConfig.Clone()
	clientTLSConfig.ServerName = "localhost"
	clientTLSConfig.RootCAs = caPool

	return serverTLSConfig, clientTLSConfig, nil
}
func TestSession_HandleAuth(t *testing.T) {
	testAuth := NewTestAuthService()
	username := "test"
	password := []byte("test@123")
	err := testAuth.AddUser(username, password)
	assert.NoError(t, err)

	serverTLSConfig, clientTLSConfig, err := getTestTLSConfig()
	assert.NoError(t, err)

	mailChan := make(chan *storage.Envelope)
	address := "localhost:20248"
	go startTesTLStServer(t, address, serverTLSConfig, mailChan, []string{"AUTH PLAIN LOGIN MD5-CRAM"}, testAuth)

	plainAuth := smtp.PlainAuth("", username, string(password), "localhost")
	md5CRAMAuth := smtp.CRAMMD5Auth(username, string(password))

	testCases := []struct {
		name string
		auth smtp.Auth
	}{
		{
			name: "Plain",
			auth: plainAuth,
		},
		{name: "MD5-CRAM",
			auth: md5CRAMAuth,
		},
	}
	for _, test := range testCases {
		c, err := smtp.Dial(address)
		assert.NoError(t, err)

		err = c.Hello("localhost")
		assert.NoError(t, err)

		err = c.StartTLS(clientTLSConfig)
		assert.NoError(t, err)

		isAuthSupported, _ := c.Extension("AUTH")
		assert.True(t, isAuthSupported)

		err = c.Auth(test.auth)
		assert.NoError(t, err)

		err = c.Mail("test0@test.com")
		assert.NoError(t, err)
		err = c.Rcpt("rtest0@test.com")
		assert.NoError(t, err)
		wc, err := c.Data()
		assert.NoError(t, err)

		err = sendMessageBody(wc, "test0@test.com", "rtest0@test.com", "Test", strings.NewReader("Hi"))
		assert.NoError(t, err)

		err = c.Quit()
		assert.NoError(t, err)

		mail := <-mailChan
		assert.Equal(t, "test0@test.com", mail.Sender)
		assert.Equal(t, "rtest0@test.com", mail.Recipient[0])
	}
}

func startTesTLStServer(t *testing.T, address string, serverTLSConfig *tls.Config, mailChan chan *storage.Envelope, exts []string, auth AuthenticationService) {

	ln, err := net.Listen("tcp", address)
	assert.NoError(t, err)
	for {
		conn, err := ln.Accept()
		assert.NoError(t, err)

		session := &Session{
			conn:       &conn,
			Conn:       textproto.NewConn(conn),
			Server:     "localhost",
			TLSConfig:  serverTLSConfig,
			Extensions: append(exts, StartTLS),
			Auth:       &auth,
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
	}
}
