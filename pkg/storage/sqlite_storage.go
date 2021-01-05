package storage

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github/ajanthan/smtp-go/pkg/smtp"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	gomail "net/mail"
	"strings"
)

type Storage interface {
	Persist(mail Mail) error
	GetAll() ([]Mail, error)
	GetBodyByMailID(mailID uint) (Content, error)
}

type SQLiteStorage struct {
	Db *gorm.DB
}

func NewStorage(dbFile string) (*SQLiteStorage, error) {
	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	if err != nil {
		return &SQLiteStorage{}, err
	}
	err = db.AutoMigrate(&Mail{}, &Content{}, &User{})
	if err != nil {
		return &SQLiteStorage{}, err
	}
	storage := &SQLiteStorage{
		Db: db,
	}
	return storage, nil
}

func (s SQLiteStorage) Persist(mail *Mail) error {
	tx := s.Db.Create(&mail)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}
func (s SQLiteStorage) GetAll() ([]Mail, error) {
	var mails []Mail
	tx := s.Db.Model(&Mail{}).Limit(10).Find(&mails)
	if tx.Error != nil {
		return mails, tx.Error
	}
	return mails, nil
}

func (s SQLiteStorage) GetBodyByMailID(mailID uint) ([]*Content, error) {
	var body []*Content
	var mail Mail
	tx := s.Db.Find(&mail, "ID=?", mailID)
	if tx.Error != nil {
		return body, tx.Error
	}

	err := s.Db.Model(&mail).Association("Body").Find(&body)
	if err != nil {
		return body, err
	}
	return body, nil
}

type DBReceiver struct {
	Storage *SQLiteStorage
}

func (p *DBReceiver) Receive(mail *smtp.Envelope) error {
	email, err := NewMail(mail.Content)
	if err != nil {
		fmt.Printf(err.Error())
	}
	return p.Storage.Persist(email)
}

type PrinterReceiver struct {
}

func (p PrinterReceiver) Receive(mail *smtp.Envelope) error {
	fmt.Println("****************************************************************************")
	fmt.Printf("From: %s\n", mail.Sender)
	fmt.Print("To:")
	for _, recipient := range mail.Recipient {
		fmt.Printf("%s,", recipient)
	}
	fmt.Println()
	email, err := NewMail(mail.Content)
	if err != nil {
		fmt.Printf(err.Error())
	}
	fmt.Printf("Message:%s", email)
	fmt.Println("****************************************************************************")
	return nil
}

func NewMail(msg *gomail.Message) (*Mail, error) {
	mail := &Mail{
		Subject:   msg.Header.Get("Subject"),
		From:      msg.Header.Get("From"),
		To:        msg.Header["To"],
		ReplyTo:   msg.Header.Get("Reply-To"),
		MessageID: msg.Header.Get("Message-ID"),
		Date:      msg.Header.Get("Date"),
	}
	err := processMailBody(msg.Body, msg.Header, mail, false, false, false)
	if err != nil {
		return nil, err
	}
	return mail, nil

}

func processMailBody(body io.Reader, headers gomail.Header, mail2 *Mail, isAttachment bool, isEmbedded bool, isAlt bool) error {
	mediaType, params, err := mime.ParseMediaType(headers.Get("Content-Type"))
	if err != nil {
		log.Fatal(err)
	}
	switch mediaType {
	case "multipart/alternative":
		isAlt = true
		mr := multipart.NewReader(body, params["boundary"])
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				log.Fatal(err)
			}
			err = processMailBody(part, gomail.Header(part.Header), mail2, isAttachment, isEmbedded, isAlt)
			if err != nil {
				return err
			}
		}

	case "multipart/related":
		isEmbedded = true
		mr := multipart.NewReader(body, params["boundary"])
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				log.Fatal(err)
			}
			err = processMailBody(part, gomail.Header(part.Header), mail2, isAttachment, isEmbedded, isAlt)
			if err != nil {
				return err
			}
		}

	case "multipart/mixed":
		isAttachment = true
		mr := multipart.NewReader(body, params["boundary"])
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				log.Fatal(err)
			}
			err = processMailBody(part, gomail.Header(part.Header), mail2, isAttachment, isEmbedded, isAlt)
			if err != nil {
				return err
			}
		}
	case "text/plain":
		fallthrough
	case "text/html":
		fallthrough
	default:
		content := &Content{}
		content.MailID = mail2.ID
		content.ContentType = headers.Get("Content-Type")
		content.Encoding = headers.Get("Content-Transfer-Encoding")
		mailBuffer := &bytes.Buffer{}
		switch strings.ToUpper(content.Encoding) {
		case "BASE64":
			_, err := mailBuffer.ReadFrom(base64.NewDecoder(base64.StdEncoding, body))
			if err != nil {
				return err
			}
		case "QUOTED-PRINTABLE":
			_, err := mailBuffer.ReadFrom(quotedprintable.NewReader(body))
			if err != nil {
				return err
			}
		case "8BIT", "7Bit":
			fallthrough
		default:
			_, err := mailBuffer.ReadFrom(body)
			if err != nil {
				return err
			}
		}
		content.Data = mailBuffer.Bytes()
		if isAlt {
			content.Type = "Alt"
			mail2.Body = append(mail2.Body, content)
		} else if isEmbedded {
			content.Type = "Emb"
			content.Name = strings.TrimRight(strings.TrimLeft(headers.Get("Content-ID"), "<"), ">")
			content.Layout = strings.Split(headers.Get("Content-Disposition"), ";")[0]
		} else if isAttachment {
			content.Type = "Att"
			parts := strings.Split(headers.Get("Content-Disposition"), ";")
			content.Layout = parts[0]
			content.Name = strings.Split(parts[1], "=")[1]
		} else {
			content.Type = "Main"
			mail2.Body = append(mail2.Body, content)
		}
	}
	return nil
}
