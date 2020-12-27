package smtp

import (
	"errors"
	"strings"
)

type Command struct {
	Name string
	Args []string
	From string
	To   string
}

func (c *Command) ParseFrom() error {
	if strings.HasPrefix(c.Args[0], "FROM:") {
		part := strings.TrimLeft(c.Args[0], "FROM:<")
		c.From = strings.TrimRight(part, ">")
		return nil
	} else {
		return errors.New("invalid MAIL command")
	}
}

func (c *Command) ParseTo() error {
	if strings.HasPrefix(c.Args[0], "TO:") {
		part := strings.TrimLeft(c.Args[0], "TO:<")
		part = strings.TrimRight(part, ">")
		c.To = part
		return nil
	} else {
		return errors.New("invalid RCPT command")
	}
}
