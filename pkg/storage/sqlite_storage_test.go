package storage

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestStorage(t *testing.T) {
	dbFile := "/tmp/testmails.db"
	t.Cleanup(func() {
		err := os.Remove(dbFile)
		if err != nil {
			t.Error(err)
		}
	})
	storage, err := NewStorage(dbFile)
	assert.NoError(t, err)
	email := &Mail{
		Subject: "Test",
		To:      []string{"test1@test.com"},
		From:    "test@test.com",
		Body: &Body{Content: &Content{
			Data: []byte("Hello!"),
		},
		},
	}
	err = storage.Persist(email)
	assert.NoError(t, err)
	mails, err := storage.GetAll()
	assert.NoError(t, err)
	body, err := storage.GetBodyByMailID(mails[0].ID)
	assert.NoError(t, err)
	assert.Equal(t, email.Subject, mails[0].Subject)
	assert.Equal(t, email.To, mails[0].To)
	assert.Equal(t, email.From, mails[0].From)
	assert.Equal(t, email.Body.Content.Data, body.Data)
	assert.Equal(t, email.Body.Content.ContentType, body.ContentType)
}
