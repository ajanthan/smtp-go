package commands

import "fmt"

type ContinueCmd struct {
	Message string
}

func (c ContinueCmd) Bytes() []byte {
	return []byte(fmt.Sprintf("354 %s\r\n", c.Message))
}
