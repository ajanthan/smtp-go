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
		Body: []*Content{{
			Data:        []byte("Hello!"),
			ContentType: "plain/text",
		},
		},
	}
	err = storage.Persist(email)
	assert.NoError(t, err)
	mails, err := storage.GetAll()
	assert.NoError(t, err)
	body, err := storage.GetBodyByMailID(mails[0].ID)
	assert.Equal(t, email.Subject, mails[0].Subject)
	assert.Equal(t, email.To, mails[0].To)
	assert.Equal(t, email.From, mails[0].From)
	assert.Equal(t, email.Body[0].Data, body[0].Data)
	assert.Equal(t, email.Body[0].ContentType, body[0].ContentType)
}
