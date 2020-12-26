package smtp

import (
	"fmt"
	"github/ajanthan/smtp-go/pkg/storage"
	"net"
	"strconv"
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
		go s.handleConnection(Connection{
			Conn: conn,
		})
	}
}

func (s Server) handleConnection(c Connection) {
	defer c.Conn.Close()

	// starting a session
	mail := storage.Envelope{}
	if err := c.Reply(STATUS_READY, s.Address+" is READY"); err != nil {
		panic(fmt.Sprintf("error sending hello %v", err))
	}

	// handling smtp commands
	for {
		cmd, err := c.ReceiveCMD()
		if err != nil {
			panic(fmt.Sprintf("error cmd %v", err))
		}
		switch cmd.Name {
		case "QUIT":
			s.Receiver.Receive(mail)
			message := fmt.Sprintf("%s service closing transmission channel", s.Address)
			if err := c.Reply(STATUS_CLOSE, message); err != nil {
				panic(fmt.Sprintf("error sending hello %v", err))
			}
			break
		case "EHLO":
			message := fmt.Sprintf("%s greets %s", s.Address, cmd.Args[0])
			if err := c.Reply(STATUS_OK, message); err != nil {
				panic(fmt.Sprintf("error sending hello %v", err))
			}
		case "HELLO":
		case "MAIL":
			sender, err := cmd.GetFrom()
			if err != nil {
				panic(fmt.Sprintf("invalid MAIL command %v", err))
			}
			mail.Sender = sender
			if err := c.Reply(STATUS_OK, "OK"); err != nil {
				panic(fmt.Sprintf("error sending hello %v", err))
			}
		case "RCPT":
			recipient, err := cmd.GetTo()
			if err != nil {
				panic(fmt.Sprintf("invalid RCPT command %v", err))
			}
			mail.Recipient = append(mail.Recipient, recipient)
			if err := c.Reply(STATUS_OK, "OK"); err != nil {
				panic(fmt.Sprintf("error sending hello %v", err))
			}
		case "DATA":
			message := "Start mail input; end with <CRLF>.<CRLF>"
			if err := c.Reply(STATUS_CONTINUE, message); err != nil {
				panic(fmt.Sprintf("error sending hello %v", err))
			}
			body, err := c.ReceiveBody()
			if err != nil {
				panic(fmt.Sprintf("error reading body %v", err))
			}
			mail.Content = body
			if err := c.Reply(STATUS_OK, "OK"); err != nil {
				panic(fmt.Sprintf("error sending hello %v", err))

			}
		}
	}
}
