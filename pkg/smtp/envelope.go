package smtp

import (
	"fmt"
	"net/mail"
	"time"
)

type Envelope struct {
	MessageID string
	Sender    string
	Recipient []string
	Content   *mail.Message
}

func NewEnvelope(serverName string) *Envelope {
	return &Envelope{MessageID: fmt.Sprintf("<%d@%s>", time.Now().Nanosecond(), serverName)}
}
