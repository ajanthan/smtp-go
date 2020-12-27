package smtp

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github/ajanthan/smtp-go/pkg/storage"
	"testing"
)

var mimeBody = ""

func TestServer_Start(t *testing.T) {
	storage := NewTestStorage()
	server := &Server{
		Address:  "localhost",
		SMTPPort: 10245,
		Storage:  storage,
		Receiver: storage,
	}
	go func() {
		server.Start()
	}()
	err := SendEmail(
		fmt.Sprintf("%s:%d", server.Address, server.SMTPPort),
		"sender@test.com",
		"receiver@test.com",
		"Test",
		"Test Message")
	assert.NoError(t, err)
	mails, err := storage.GetAll()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(mails))
	assert.Equal(t, "sender@test.com", mails[0].From)
	assert.Equal(t, "receiver@test.com", mails[0].To[0])
	assert.Equal(t, "Test", mails[0].Subject)
	assert.Equal(t, "Test Message", string(mails[0].Body.Data))
	//"../../resources/mime_body.txt"
	err = SendEmailFromFile(
		fmt.Sprintf("%s:%d", server.Address, server.SMTPPort),
		"test <wso2iamtest@gmail.com>",
		"subash@wso2.com",
		"../../resources/mime_body.txt")
	assert.NoError(t, err)

	mails, err = storage.GetAll()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(mails))
	assert.Equal(t, "test <wso2iamtest@gmail.com>", mails[1].From)
	assert.Equal(t, "subash@wso2.com", mails[1].To[0])
	assert.Equal(t, "WSO2 - Password Reset", mails[1].Subject)
	assert.Equal(t, "text/html; charset=UTF-8", mails[1].Body.ContentType)
	assert.NotNil(t, mails[1].Body.Data)
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
func (t *TestStorage) Receive(mail storage.Envelope) error {
	email, err := parseMail(mail.Content)
	if err != nil {
		return err
	}
	return t.Persist(email)
}
