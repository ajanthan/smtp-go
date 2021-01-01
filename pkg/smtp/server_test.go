package smtp

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/mail"
	"strings"
	"testing"
)

//TODO: Add benchmark tests
func TestNone_MIME_Mail(t *testing.T) {
	testStorage := NewTestStorage()
	server := &Server{
		Address:  "localhost",
		SMTPPort: 10245,
		Receiver: testStorage,
	}
	go func() {
		server.Start()
	}()
	err := SendEmail(
		fmt.Sprintf("%s:%d", server.Address, server.SMTPPort),
		"sender@test.com",
		"receiver@test.com",
		"Test",
		strings.NewReader("Test Message\n"))
	require.NoError(t, err)
	mails, err := testStorage.GetAll()
	require.NoError(t, err)
	assert.Equal(t, 1, len(mails))
	assert.Equal(t, "sender@test.com", mails[0].Sender)
	assert.Equal(t, "receiver@test.com", mails[0].Recipient[0])
	assert.Equal(t, "Test", mails[0].Content.Header.Get("Subject"))
	buffer := bytes.Buffer{}
	_, err = buffer.ReadFrom(mails[0].Content.Body)
	require.NoError(t, err)
	assert.Equal(t, "Test Message\n", buffer.String())
}

func TestMiME_Mail(t *testing.T) {
	testStorage := NewTestStorage()
	server := &Server{
		Address:  "localhost",
		SMTPPort: 10246,
		Receiver: testStorage,
	}
	go func() {
		server.Start()
	}()
	//"../../resources/mime_body.txt"
	err := SendEmailFromFile(
		fmt.Sprintf("%s:%d", server.Address, server.SMTPPort),
		"wso2iamtest@gmail.com",
		"subash@wso2.com",
		"../../resources/mime_body.txt")
	require.NoError(t, err)

	mails, err := testStorage.GetAll()
	require.NoError(t, err)
	assert.Equal(t, 1, len(mails))
	assert.Equal(t, "wso2iamtest@gmail.com", mails[0].Sender)
	assert.Equal(t, "subash@wso2.com", mails[0].Recipient[0])
	assert.Equal(t, "WSO2 - Password Reset", mails[0].Content.Header.Get("Subject"))
	assert.Equal(t, "text/html; charset=UTF-8", mails[0].Content.Header.Get("Content-Type"))
	assert.NotNil(t, mails[0].Content)
}

type TestStorage struct {
	mails     map[uint]Envelope
	idCounter uint
}

func NewTestStorage() *TestStorage {
	mails := make(map[uint]Envelope)
	return &TestStorage{
		mails: mails,
	}
}
func (t *TestStorage) Persist(mail Envelope) error {
	t.mails[t.idCounter] = mail
	t.idCounter++
	return nil
}
func (t *TestStorage) GetAll() ([]Envelope, error) {
	var mails []Envelope
	for _, mail := range t.mails {
		mails = append(mails, mail)
	}
	return mails, nil
}
func (t *TestStorage) GetBodyByMailID(mailID uint) (mail.Message, error) {
	return *t.mails[mailID].Content, nil
}
func (t *TestStorage) Receive(mail *Envelope) error {
	return t.Persist(*mail)
}
