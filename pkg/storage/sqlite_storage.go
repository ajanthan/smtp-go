package storage

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Storage interface {
	Persist(mail Mail) error
	GetAll() ([]Mail, error)
	GetBodyByMailID(mailID uint) (Body, error)
}

type SQLiteStorage struct {
	Db *gorm.DB
}

func NewStorage(dbFile string) (*SQLiteStorage, error) {
	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	if err != nil {
		return &SQLiteStorage{}, err
	}
	err = db.AutoMigrate(&Mail{}, &Body{})
	if err != nil {
		return &SQLiteStorage{}, err
	}
	storage := &SQLiteStorage{
		Db: db,
	}
	return storage, nil
}

func (s SQLiteStorage) Persist(mail Mail) error {
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

func (s SQLiteStorage) GetBodyByMailID(mailID uint) (Body, error) {
	var body Body
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
