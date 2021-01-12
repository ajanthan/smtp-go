package storage

import (
	"bytes"
	"encoding/base64"
	"errors"
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
	err = db.AutoMigrate(&Mail{}, &Body{}, &Attachment{}, &EmbeddedFile{}, &Alternative{}, &User{})
	if err != nil {
		return &SQLiteStorage{}, err
	}
	storage := &SQLiteStorage{
		Db: db,
	}
	return storage, nil
}

func (s SQLiteStorage) Persist(mail *Mail) error {
	tx := s.Db.Create(mail)
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

func (s SQLiteStorage) GetBodyByMailID(mailID uint) (*Body, error) {
	var body Body
	var mail Mail
	tx := s.Db.Find(&mail, "ID=?", mailID)
	if tx.Error != nil {
		return &body, tx.Error
	}

	err := s.Db.Model(&mail).Association("Body").Find(&body)
	if err != nil {
		return &body, err
	}
	return &body, nil
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
	mediaType, params, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}
	switch mediaType {
	case "multipart/mixed":
		multiPartMail, err := processMultipartMixed(params["boundary"], msg.Body)
		if err != nil {
			return nil, err
		}
		mail.Body = multiPartMail.Body
		mail.Alternatives = multiPartMail.Alternatives
		mail.Attachments = multiPartMail.Attachments
	case "multipart/related":
		body, err := processMultipartRelated(params["boundary"], msg.Body)
		if err != nil {
			return nil, err
		}
		mail.Body = body
	case "multipart/alternative":
		bodies, err := processMultipartAlternative(params["boundary"], msg.Body)
		if err != nil {
			return nil, err
		}
		if len(bodies) > 0 {
			mail.Body = bodies[0]
			if len(bodies) > 1 {
				for _, alternative := range bodies[1:] {
					mail.Alternatives = append(mail.Alternatives, &Alternative{Content: alternative.Content})
				}
			}
		}
	case "text/plain", "text/html":
		content, err := processMailContent(msg.Body, msg.Header)
		if err != nil {
			return nil, err
		}
		mail.Body = &Body{Content: content}
	default:
		return nil, errors.New("unsupported content type")
	}
	return mail, nil
}

func processMultipartMixed(boundary string, body io.Reader) (*Mail, error) {
	mr := multipart.NewReader(body, boundary)
	mail := &Mail{}
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		mediaType, params, err := mime.ParseMediaType(part.Header.Get("Content-Type"))
		if err != nil {
			return nil, err
		}
		switch mediaType {
		case "multipart/related":
			mail.Body, err = processMultipartRelated(params["boundary"], part)
			if err != nil {
				return nil, err
			}

		case "multipart/alternative":
			bodies, err := processMultipartAlternative(params["boundary"], part)
			if err != nil {
				return nil, err
			}
			if len(bodies) > 0 {
				mail.Body = bodies[0]
				if len(bodies) > 1 {
					for _, alternative := range bodies[1:] {
						mail.Alternatives = append(mail.Alternatives, &Alternative{Content: alternative.Content})
					}
				}
			}
		case "text/plain", "text/html":
			content, err := processMailContent(part, gomail.Header(part.Header))
			if err != nil {
				return nil, err
			}
			mail.Body = &Body{Content: content}
		default:
			content, err := processMailContent(part, gomail.Header(part.Header))
			if err != nil {
				return nil, err
			}
			mail.Attachments = append(mail.Attachments, &Attachment{Content: content})
		}
	}
	return mail, nil
}
func processMultipartRelated(boundary string, body io.Reader) (*Body, error) {
	mr := multipart.NewReader(body, boundary)
	mailBody := &Body{}
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if !isMiMEBody(gomail.Header(part.Header)) {
			content, err := processMailContent(part, gomail.Header(part.Header))
			if err != nil {
				return nil, err
			}
			if content.ContentType == "plain/text" || content.ContentType == "plain/html" {
				mailBody.Content = content
			} else {
				mailBody.Embeds = append(mailBody.Embeds, &EmbeddedFile{Content: content})
			}
		} else {
			return nil, errors.New("multipart content inside multipart/related")
		}
	}
	return mailBody, nil
}
func processMultipartAlternative(boundary string, body io.Reader) ([]*Body, error) {
	var alternativeContents []*Body
	mr := multipart.NewReader(body, boundary)
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if isMiMEBody(gomail.Header(part.Header)) && part.Header.Get("Content-Type") == "multipart/related" {
			body, err := processMultipartRelated(boundary, part)
			if err != nil {
				return nil, err
			}
			alternativeContents = append(alternativeContents, body)
		} else {
			content, err := processMailContent(part, gomail.Header(part.Header))
			body := &Body{Content: content}
			if err != nil {
				return nil, err
			}
			alternativeContents = append(alternativeContents, body)
		}
	}
	return alternativeContents, nil
}
func isMiMEBody(headers gomail.Header) bool {
	mediaType, _, err := mime.ParseMediaType(headers.Get("Content-Type"))
	if err != nil {
		return false
	}
	if strings.HasPrefix(mediaType, "multipart/") {
		return true
	}
	return false
}
func processMailContent(body io.Reader, headers gomail.Header) (*Content, error) {
	content := &Content{}
	content.ContentType = headers.Get("Content-Type")
	content.Encoding = headers.Get("Content-Transfer-Encoding")
	mailBuffer := &bytes.Buffer{}
	switch strings.ToUpper(content.Encoding) {
	case "BASE64":
		_, err := mailBuffer.ReadFrom(base64.NewDecoder(base64.StdEncoding, body))
		if err != nil {
			return nil, err
		}
	case "QUOTED-PRINTABLE":
		_, err := mailBuffer.ReadFrom(quotedprintable.NewReader(body))
		if err != nil {
			return nil, err
		}
	case "8BIT", "7Bit":
		fallthrough
	default:
		_, err := mailBuffer.ReadFrom(body)
		if err != nil {
			return nil, err
		}
	}
	content.Data = mailBuffer.Bytes()
	contentDepHeader := headers.Get("Content-Disposition")
	if contentDepHeader != "" {
		displayType, params, err := parseContentDepositionHeader(headers.Get("Content-Disposition"))
		if err != nil {
			return nil, err
		}
		if displayType == "attachment" {
			content.Type = "attachment"
			filename, ok := params["filename"]
			if !ok {
				_, mParams, err := mime.ParseMediaType(headers.Get("Content-Type"))
				if err != nil {
					return nil, err
				}
				filename, ok = mParams["name"]
				if !ok {
					return nil, errors.New("unable to figure out attachment name")
				}
			}
			content.Name = filename
			return content, nil
		} else if displayType == "inline" {
			content.Type = "inline"
			content.Name = headers.Get("Content-ID")
			return content, nil
		}
	}
	return content, nil
}

func parseContentDepositionHeader(header string) (string, map[string]string, error) {
	displayType := ""
	params := make(map[string]string)
	contentDepParts := strings.Split(header, ";")
	if len(contentDepParts) > 0 {
		displayType = contentDepParts[0]
		if len(contentDepParts) > 1 {
			for _, contentDepPart := range contentDepParts[1:] {
				paramParts := strings.Split(contentDepPart, "=")
				if len(paramParts) > 1 {
					params[paramParts[0]] = paramParts[1]
				}
			}
		}
		return displayType, params, nil
	}
	return "", nil, errors.New("invalid Content-Disposition header")
}
