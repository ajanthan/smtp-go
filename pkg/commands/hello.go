package commands

import "fmt"

type HelloCmd struct {
	Domain    string
	Greetings string
}

func (h HelloCmd) Bytes() []byte {
	return []byte(fmt.Sprintf("220 HELLO %s %s \r\n", h.Domain, h.Greetings))
}
