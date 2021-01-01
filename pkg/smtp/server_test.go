package smtp

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github/ajanthan/smtp-go/pkg/storage"
	"strings"
	"testing"
)

//TODO: Add benchmark tests
func TestNone_MIME_Mail(t *testing.T) {
	testStorage := NewTestStorage()
	server := &Server{
		Address:  "localhost",
		SMTPPort: 10245,
		Storage:  testStorage,
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
	assert.NoError(t, err)
	mails, err := testStorage.GetAll()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(mails))
	assert.Equal(t, "sender@test.com", mails[0].From)
	assert.Equal(t, "receiver@test.com", mails[0].To[0])
	assert.Equal(t, "Test", mails[0].Subject)
	assert.Equal(t, "Test Message\n", string(mails[0].Body.Data))
}

func TestMiME_Mail(t *testing.T) {
	testStorage := NewTestStorage()
	server := &Server{
		Address:  "localhost",
		SMTPPort: 10246,
		Storage:  testStorage,
		Receiver: testStorage,
	}
	go func() {
		server.Start()
	}()
	//"../../resources/mime_body.txt"
	err := SendEmailFromFile(
		fmt.Sprintf("%s:%d", server.Address, server.SMTPPort),
		"test <wso2iamtest@gmail.com>",
		"subash@wso2.com",
		"../../resources/mime_body.txt")
	assert.NoError(t, err)

	mails, err := testStorage.GetAll()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(mails))
	assert.Equal(t, "test <wso2iamtest@gmail.com>", mails[0].From)
	assert.Equal(t, "subash@wso2.com", mails[0].To[0])
	assert.Equal(t, "WSO2 - Password Reset", mails[0].Subject)
	assert.Equal(t, "text/html; charset=UTF-8", mails[0].Body.ContentType)
	assert.NotNil(t, mails[0].Body.Data)
}

type TestStorage struct {
	mails     map[uint]storage.Mail
	idCounter uint
}

func NewTestStorage() *TestStorage {
	mails := make(map[uint]storage.Mail)
	return &TestStorage{
		mails: mails,
	}
}
func (t *TestStorage) Persist(mail storage.Mail) error {
	t.mails[t.idCounter] = mail
	t.idCounter++
	return nil
}
func (t *TestStorage) GetAll() ([]storage.Mail, error) {
	var mails []storage.Mail
	for _, mail := range t.mails {
		mails = append(mails, mail)
	}
	return mails, nil
}
func (t *TestStorage) GetBodyByMailID(mailID uint) (storage.Body, error) {
	return t.mails[mailID].Body, nil
}
func (t *TestStorage) Receive(mail *storage.Envelope) error {
	email, err := newMail(mail.Content)
	if err != nil {
		return err
	}
	return t.Persist(email)
}
