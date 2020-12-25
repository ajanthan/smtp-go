package storage

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"strings"
)

type Envelope struct {
	Sender    string
	Recipient []string
	Content   []byte
}

type Mail struct {
	gorm.Model
	Date      string
	From      string
	ReplyTo   string
	Subject   string
	MessageID string
	To        Recipients `sql:"type:text"`
	Body      Body
}

type Recipients []string

func (r Recipients) Value() (driver.Value, error) {
	bytes, err := json.Marshal(r)
	return string(bytes), err
}

func (r *Recipients) Scan(input interface{}) error {
	switch value := input.(type) {
	case string:
		return json.Unmarshal([]byte(value), r)
	case []byte:
		return json.Unmarshal(value, r)
	default:
		return errors.New("unsupported type")
	}
}

type Body struct {
	gorm.Model
	MailID      uint
	Data        []byte
	ContentType string
}

func (b Body) String() string {
	var builder strings.Builder
	builder.WriteString("{\n")
	builder.WriteString("Data:" + string(b.Data) + ",\n")
	builder.WriteString("Content-Type:" + b.ContentType + ",\n")
	builder.WriteString("}\n")
	return builder.String()
}

func (m Mail) String() string {
	var builder strings.Builder
	builder.WriteString("{\n")
	builder.WriteString("Subject:" + m.Subject + ",\n")
	builder.WriteString("From:" + m.From + ",\n")
	builder.WriteString("To:" + fmt.Sprintf("%v", m.To) + ",\n")
	builder.WriteString("ReplyTo:" + m.ReplyTo + ",\n")
	builder.WriteString("MessageID:" + m.MessageID + ",\n")
	builder.WriteString("Date:" + m.Date + ",\n")
	builder.WriteString("Body:" + m.Body.String() + "\n")
	builder.WriteString("}\n")
	return builder.String()

}
