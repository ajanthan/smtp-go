package smtp

import (
	"errors"
	"strings"
)

type Command struct {
	Name string
	Args []string
}

func (c Command) GetFrom() (string, error) {
	if strings.HasPrefix(c.Args[0], "FROM:") {
		part := strings.TrimLeft(c.Args[0], "FROM:<")
		return strings.TrimRight(part, ">"), nil
	} else {
		return "", errors.New("invalid MAIL command")
	}
}

func (c Command) GetTo() (string, error) {
	if strings.HasPrefix(c.Args[0], "TO:") {
		part := strings.TrimLeft(c.Args[0], "TO:<")
		part = strings.TrimRight(part, ">")
		return part, nil
	} else {
		return "", errors.New("invalid RCPT command")
	}
}
