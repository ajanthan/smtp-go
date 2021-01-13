package storage

import (
	"fmt"
	"github/ajanthan/smtp-go/pkg/smtp"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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
	//mail.Content.Body = io.TeeReader(mail.Content.Body, os.Stdout)
	email, err := NewMail(mail.Content)
	if err != nil {
		return err
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
