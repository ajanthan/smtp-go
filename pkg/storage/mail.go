package storage

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"strings"
)

type Mail struct {
	gorm.Model
	Date         string
	From         string
	ReplyTo      string
	Subject      string
	MessageID    string
	To           Recipients `sql:"type:text"`
	Body         *Body
	Alternatives []*Alternative
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

type Attachment struct {
	gorm.Model
	*Content
	BodyID uint
}
type Alternative struct {
	*Body
}
type EmbeddedFile struct {
	gorm.Model
	*Content
	BodyID uint
}
type Body struct {
	gorm.Model
	*Content
	Embeds      []*EmbeddedFile
	Attachments []*Attachment
	MailID      uint
}

type Content struct {
	Data        []byte
	ContentType string
	Encoding    string
	Type        string
	Layout      string
	Name        string
}

func (b *Content) String() string {
	var builder strings.Builder
	builder.WriteString("{\n")
	builder.WriteString("Content:" + string(b.Data) + ",\n")
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
	builder.WriteString("Body:" + fmt.Sprintf("%v", m.Body) + "\n")
	builder.WriteString("}\n")
	return builder.String()

}
