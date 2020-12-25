package commands

import "fmt"

type Ok struct {
	Message string
}

func (ok Ok) Bytes() []byte {
	return []byte(fmt.Sprintf("250 %s\r\n", ok.Message))
}
