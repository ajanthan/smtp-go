package smtp

import (
	"fmt"
	"github/ajanthan/smtp-go/pkg/commands"
	"github/ajanthan/smtp-go/pkg/storage"
	"net"
	"strconv"
	"strings"
)

type Server struct {
	Address  string
	SMTPPort int
	Receiver MailReceiver
	Storage  storage.Storage
}

func (s Server) Start() {
	ln, err := net.Listen("tcp", s.Address+":"+strconv.Itoa(s.SMTPPort))
	if err != nil {
		panic(fmt.Sprintf("error starting server %v", err))
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(fmt.Sprintf("error accepting conn %v", err))
		}
		go s.handleConnection(conn)
	}
}

func (s Server) handleConnection(c net.Conn) {
	defer c.Close()
	//starting a session
	mail := storage.Envelope{}
	helloCmd := commands.HelloCmd{
		Domain:    s.Address,
		Greetings: "is READY",
	}
	if _, err := c.Write(helloCmd.Bytes()); err != nil {
		panic(fmt.Sprintf("error sending hello %v", err))

	}
	for {
		buff := make([]byte, 1024)
		if _, err := c.Read(buff); err != nil {
			panic(fmt.Sprintf("error reading from conn %v", err))

		}
		cmdIn := string(buff)
		cmdIn = strings.Split(cmdIn, "\r\n")[0]
		args := strings.Split(cmdIn, " ")
		cmd := args[0]
		args = args[1:len(args)]
		switch cmd {
		case "QUIT":
			quitCmd := commands.QuitCmd{
				Message: fmt.Sprintf("%s service closing transmission channel", s.Address),
			}
			s.Receiver.Receive(mail)
			if _, err := c.Write(quitCmd.Bytes()); err != nil {
				panic(fmt.Sprintf("error sending hello %v", err))

			}
			return
		case "EHLO":
			okCmd := commands.Ok{
				Message: s.Address,
			}
			if _, err := c.Write(okCmd.Bytes()); err != nil {
				panic(fmt.Sprintf("error sending hello %v", err))

			}
		case "HELLO":
		case "MAIL":
			if strings.HasPrefix(args[0], "FROM:") {
				part := strings.TrimLeft(args[0], "FROM:<")
				mail.Sender = strings.TrimRight(part, ">")
			}
			okCmd := commands.Ok{
				Message: "OK",
			}
			if _, err := c.Write(okCmd.Bytes()); err != nil {
				panic(fmt.Sprintf("error sending hello %v", err))

			}
		case "RCPT":
			if strings.HasPrefix(args[0], "TO:") {
				part := strings.TrimLeft(args[0], "TO:<")
				part = strings.TrimRight(part, ">")
				mail.Recipient = append(mail.Recipient, part)
			}
			okCmd := commands.Ok{
				Message: "OK",
			}
			if _, err := c.Write(okCmd.Bytes()); err != nil {
				panic(fmt.Sprintf("error sending hello %v", err))

			}
		case "DATA":
			continueCmd := commands.ContinueCmd{
				Message: "Start mail input; end with <CRLF>.<CRLF>",
			}
			if _, err := c.Write(continueCmd.Bytes()); err != nil {
				panic(fmt.Sprintf("error sending hello %v", err))

			}
			message, err := readBody(c)
			if err != nil {
				panic(fmt.Sprintf("error reading body %v", err))
			}
			mail.Content = message
			okCmd := commands.Ok{
				Message: "OK",
			}
			if _, err := c.Write(okCmd.Bytes()); err != nil {
				panic(fmt.Sprintf("error sending hello %v", err))

			}
		}
	}
}

func readBody(c net.Conn) ([]byte, error) {
	var message []byte
	for {
		buff := make([]byte, 1024)
		if _, err := c.Read(buff); err != nil {
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
