package storage

import (
	"crypto/hmac"
	"crypto/md5"
	"fmt"
	"github/ajanthan/smtp-go/pkg/smtp"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string
	Password []byte
}

func (s *SQLiteStorage) Authenticate(username string, password []byte) error {
	var user User
	tx := s.Db.Where("username=?", username).Find(&user)
	if tx.Error != nil {
		return tx.Error
	}
	err := bcrypt.CompareHashAndPassword(user.Password, password)
	if err != nil {
		return smtp.NewInvalidCredentialError("invalid credential")
	}
	return nil
}
func (s *SQLiteStorage) ValidateHMAC(username string, msg []byte, code []byte) error {
	var user User
	tx := s.Db.Where("username=?", username).Find(&user)
	if tx.Error != nil {
		return tx.Error
	}
	hmacEncoder := hmac.New(md5.New, user.Password)
	hmacEncoder.Write(msg)
	expectedCode := make([]byte, 0, hmacEncoder.Size())
	expectedCode = hmacEncoder.Sum(expectedCode)
	expectedCodeStr := []byte(fmt.Sprintf("%x", expectedCode))

	if len(expectedCodeStr) != len(code) {
		return smtp.NewInvalidCredentialError("invalid credential")
	}
	for i, b := range code {
		if b != expectedCodeStr[i] {
			return smtp.NewInvalidCredentialError("invalid credential")
		}
	}
	return nil
}
func (s *SQLiteStorage) AddUser(username string, password []byte) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword(password, 10)
	if err != nil {
		return "", err
	}
	tx := s.Db.Create(&User{
		Username: username,
		Password: hashedPassword,
	})
	if tx.Error != nil {
		return "", tx.Error
	}
	return string(hashedPassword), nil
}
