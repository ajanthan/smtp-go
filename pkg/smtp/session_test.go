package smtp

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net"
	"net/smtp"
	"net/textproto"
	"strings"
	"testing"
)

func TestSession_HandleReset(t *testing.T) {
	mailChan := make(chan *Envelope)
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
	require.NoError(t, err)
	err = c.Mail("test0@test.com")
	require.NoError(t, err)
	err = c.Rcpt("rtest0@test.com")
	require.NoError(t, err)
	err = c.Reset()
	require.NoError(t, err)
	err = c.Mail("test1@test.com")
	require.NoError(t, err)
	err = c.Rcpt("rtest1@test.com")
	require.NoError(t, err)
	wc, err := c.Data()
	require.NoError(t, err)

	err = sendMessageBody(wc, "test1@test.com", "rtest1@test.com", "Test", strings.NewReader("Hi"))
	require.NoError(t, err)

	err = c.Quit()
	require.NoError(t, err)

	mail := <-mailChan
	assert.Equal(t, "test1@test.com", mail.Sender)
	assert.Equal(t, "rtest1@test.com", mail.Recipient[0])
}

func TestSession_HandleStartTLS(t *testing.T) {
	serverTLSConfig, clientTLSConfig, err := getTestTLSConfig()
	require.NoError(t, err)

	mailChan := make(chan *Envelope)
	address := "localhost:20247"
	go startTesTLStServer(t, address, serverTLSConfig, mailChan, []string{}, nil, false)

	c, err := smtp.Dial(address)
	require.NoError(t, err)

	err = c.Hello("localhost")
	require.NoError(t, err)

	isTLSSupported, _ := c.Extension("STARTTLS")
	assert.True(t, isTLSSupported)

	err = c.StartTLS(clientTLSConfig)
	require.NoError(t, err)

	err = c.Mail("test0@test.com")
	require.NoError(t, err)
	err = c.Rcpt("rtest0@test.com")
	require.NoError(t, err)
	err = c.Reset()
	require.NoError(t, err)
	err = c.Mail("test1@test.com")
	require.NoError(t, err)
	err = c.Rcpt("rtest1@test.com")
	require.NoError(t, err)
	wc, err := c.Data()
	require.NoError(t, err)

	err = sendMessageBody(wc, "test1@test.com", "rtest1@test.com", "Test", strings.NewReader("Hi"))
	require.NoError(t, err)

	err = c.Quit()
	require.NoError(t, err)

	mail := <-mailChan
	assert.Equal(t, "test1@test.com", mail.Sender)
	assert.Equal(t, "rtest1@test.com", mail.Recipient[0])
}

func TestSession_HandleAuth(t *testing.T) {
	testAuth := NewTestAuthService()
	username := "test"
	password := []byte("test@123")
	err := testAuth.AddUser(username, password)
	require.NoError(t, err)

	serverTLSConfig, clientTLSConfig, err := getTestTLSConfig()
	require.NoError(t, err)

	mailChan := make(chan *Envelope)
	address := "localhost:20248"
	go startTesTLStServer(t, address, serverTLSConfig, mailChan, []string{"AUTH PLAIN LOGIN MD5-CRAM"}, testAuth, true)

	testCases := []struct {
		name      string
		checkAuth func(t *testing.T, c *smtp.Client) bool
	}{
		{
			name: "Plain",
			checkAuth: func(t *testing.T, c *smtp.Client) bool {
				plainAuth := smtp.PlainAuth("", username, string(password), "localhost")
				err = c.Auth(plainAuth)
				require.NoError(t, err)
				return true
			},
		},
		{
			name: "MD5-CRAM",
			checkAuth: func(t *testing.T, c *smtp.Client) bool {
				md5CRAMAuth := smtp.CRAMMD5Auth(username, string(password))
				err = c.Auth(md5CRAMAuth)
				require.NoError(t, err)
				return true
			},
		},
		{
			name: "Negative",
			checkAuth: func(t *testing.T, c *smtp.Client) bool {
				//No auth
				//Expect error
				err = c.Mail("test0@test.com")
				require.Error(t, err)
				err = c.Quit()
				require.NoError(t, err)
				<-mailChan
				return false
			},
		},
		{
			name: "wrong credential",
			checkAuth: func(t *testing.T, c *smtp.Client) bool {
				plainAuth := smtp.PlainAuth("", username, "pawn", "localhost")
				err = c.Auth(plainAuth)
				require.Error(t, err)
				<-mailChan
				return false
			},
		},
	}
	for _, test := range testCases {
		c, err := smtp.Dial(address)
		require.NoError(t, err)

		err = c.Hello("localhost")
		require.NoError(t, err)

		err = c.StartTLS(clientTLSConfig)
		require.NoError(t, err)

		isAuthSupported, _ := c.Extension("AUTH")
		assert.True(t, isAuthSupported)

		if test.checkAuth(t, c) {
			err = c.Mail("test0@test.com")
			require.NoError(t, err)
			err = c.Rcpt("rtest0@test.com")
			require.NoError(t, err)
			wc, err := c.Data()
			require.NoError(t, err)

			err = sendMessageBody(wc, "test0@test.com", "rtest0@test.com", "Test", strings.NewReader("Hi"))
			require.NoError(t, err)

			err = c.Quit()
			require.NoError(t, err)

			mail := <-mailChan
			assert.Equal(t, "test0@test.com", mail.Sender)
			assert.Equal(t, "rtest0@test.com", mail.Recipient[0])
		}
	}
}

func startTesTLStServer(t *testing.T, address string, serverTLSConfig *tls.Config, mailChan chan *Envelope, exts []string, auth AuthenticationService, isSecure bool) {
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
			Auth:       auth,
			Secure:     isSecure,
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
			assert.NoError(t, err)
		}
		mailChan <- mail
	}
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
