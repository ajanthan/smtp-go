package smtp

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

type Connection struct {
	Conn net.Conn
}

func (c Connection) Reply(statusCode int, statusLine string) error {
	if _, err := c.Conn.Write([]byte(fmt.Sprintf("%d %s\r\n", statusCode, statusLine))); err != nil {
		return err
	}
	return nil
}

func (c Connection) ReceiveCMD() (Command, error) {
	buff := make([]byte, 1024)
	command := Command{}
	if _, err := c.Conn.Read(buff); err != nil {
		return Command{}, err
	}

	cmdIn := strings.Split(string(buff), "\r\n")
	if len(cmdIn) != 2 {
		return command, errors.New("invalid command format")
	}
	args := strings.Split(cmdIn[0], " ")
	command.Name = args[0]
	command.Args = args[1:]
	return command, nil
}
func (c Connection) ReceiveBody() ([]byte, error) {
	var message []byte
	for {
		buff := make([]byte, 1024)
		if _, err := c.Conn.Read(buff); err != nil {
			return nil, err
		}
		if strings.Contains(string(buff), "\r\n.\r\n") {
			line := strings.Split(string(buff), "\r\n.\r\n")
			message = append(message, line[0]...)
			return message, nil
		}
		message = append(message, buff...)
	}
}
