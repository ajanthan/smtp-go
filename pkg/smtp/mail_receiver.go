package smtp

import (
	"bytes"
	"fmt"
	"github/ajanthan/smtp-go/pkg/storage"
	"strings"
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
	email, err := parseMail(mail.Content)
	if err != nil {
		fmt.Printf(err.Error())
	}
	fmt.Printf("Message:%s", email)
	fmt.Println("****************************************************************************")
	return nil
}

func parseMail(body []byte) (storage.Mail, error) {
	parts := bytes.SplitN(body, []byte("\r\n\r\n"), 2)
	headerParts := strings.Split(string(parts[0]), "\r\n")
	headers := make(map[string][]string)
	for _, headerPart := range headerParts {
		keyVal := strings.SplitN(headerPart, ":", 2)
		headers[keyVal[0]] = keyVal[1:]
	}
	mail := storage.Mail{
		Subject:   headers["Subject"][0],
		From:      headers["From"][0],
		To:        headers["To"],
		ReplyTo:   headers["Reply-To"][0],
		MessageID: headers["Message-ID"][0],
		Date:      headers["Date"][0],
		Body: storage.Body{
			ContentType: headers["Content-Type"][0],
			Data:        parts[1]},
	}
	return mail, nil

}

type DBReceiver struct {
	Storage *storage.Storage
}

func (p *DBReceiver) Receive(mail storage.Envelope) error {
	email, err := parseMail(mail.Content)
	if err != nil {
		fmt.Printf(err.Error())
	}
	return p.Storage.Persist(email)
}
