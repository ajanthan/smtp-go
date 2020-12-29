package smtp

import (
	"bytes"
	"fmt"
	"github/ajanthan/smtp-go/pkg/storage"
	mail2 "net/mail"
)

type MailReceiver interface {
	Receive(mail storage.Envelope) error
}

type PrinterReceiver struct {
}

func (p PrinterReceiver) Receive(mail storage.Envelope) error {
	fmt.Println("****************************************************************************")
	fmt.Printf("From: %s\n", mail.Sender)
	fmt.Print("To:")
	for _, recipient := range mail.Recipient {
		fmt.Printf("%s,", recipient)
	}
	fmt.Println()
	email, err := newMail(mail.Content)
	if err != nil {
		fmt.Printf(err.Error())
	}
	fmt.Printf("Message:%s", email)
	fmt.Println("****************************************************************************")
	return nil
}

func newMail(msg *mail2.Message) (storage.Mail, error) {
	mail := storage.Mail{
		Subject:   msg.Header.Get("Subject"),
		From:      msg.Header.Get("From"),
		To:        msg.Header["To"],
		ReplyTo:   msg.Header.Get("Reply-To"),
		MessageID: msg.Header.Get("Message-ID"),
		Date:      msg.Header.Get("Date"),
	}
	_, isMIME := msg.Header["Mime-Version"]
	if isMIME {
		mail.Body = storage.Body{
			ContentType: msg.Header.Get("Content-Type")}
	} else {
		mail.Body = storage.Body{}
	}
	buff := new(bytes.Buffer)
	_, err := buff.ReadFrom(msg.Body)
	if err != nil {
		return storage.Mail{}, err
	}
	mail.Body.Data = buff.Bytes()
	return mail, nil

}

type DBReceiver struct {
	Storage *storage.SQLiteStorage
}

func (p *DBReceiver) Receive(mail storage.Envelope) error {
	email, err := newMail(mail.Content)
	if err != nil {
		fmt.Printf(err.Error())
	}
	return p.Storage.Persist(email)
}
