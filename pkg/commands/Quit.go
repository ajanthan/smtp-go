package commands

import "fmt"

type QuitCmd struct {
	Message string
}

func (q QuitCmd) Bytes() []byte {
	return []byte(fmt.Sprintf("221 %s\r\n", q.Message))
}
