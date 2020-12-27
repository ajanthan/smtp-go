package smtp

import (
	"errors"
	"fmt"
	"github/ajanthan/smtp-go/pkg/storage"
	"log"
	"net"
	"strings"
)

type Session struct {
	Conn                     net.Conn
	Server                   string
	Client                   string
	IsHelloReceived          bool
	IsMailReceived           bool
	IsAtLeastOneRcptReceived bool
}

func (s *Session) Start() error {
	greetings := s.Server + " is READY"
	if err := s.Reply(StatusReady, greetings); err != nil {
		return NewServerError(fmt.Sprintf("error sending ready message %v", err))
	}
	return nil
}

func (s *Session) HandleHello(cmd Command) error {
	if s.IsHelloReceived {
		return NewOutOfOrderCmdError(fmt.Sprintf("%s is already received", cmd.Name))
	}
	s.Client = cmd.Args[0]
	message := fmt.Sprintf("%s greets %s", s.Server, s.Client)
	if err := s.Reply(StatusOk, message); err != nil {
		return NewServerError(fmt.Sprintf("error sending hello %v", err))
	}
	s.IsHelloReceived = true
	return nil
}
func (s *Session) HandleMail(cmd Command) (string, error) {
	if !s.IsHelloReceived {
		return "", NewOutOfOrderCmdError("MAIL command before EHLO/HELLO command")
	} else if s.IsMailReceived {
		return "", NewOutOfOrderCmdError("MAIL command is already received")
	}
	if err := s.Reply(StatusOk, "OK"); err != nil {
		return "", NewServerError(fmt.Sprintf("error sending hello %v", err))
	}
	s.IsMailReceived = true
	return cmd.From, nil
}
func (s *Session) HandleRcpt(cmd Command) (string, error) {
	if !s.IsHelloReceived {
		return "", NewOutOfOrderCmdError("RCPT command before EHLO/HELLO command")
	} else if !s.IsMailReceived {
		return "", NewOutOfOrderCmdError("RCPT command before MAIL command")
	}
	if !s.IsAtLeastOneRcptReceived {
		s.IsAtLeastOneRcptReceived = true
	}
	if err := s.Reply(StatusOk, "OK"); err != nil {
		return "", NewServerError(fmt.Sprintf("error sending hello %v", err))
	}
	return cmd.To, nil
}
func (s *Session) HandleData(Command) ([]byte, error) {
	if !s.IsHelloReceived {
		return []byte{}, NewOutOfOrderCmdError("DATA command before EHLO/HELLO command")
	} else if !s.IsMailReceived {
		return []byte{}, NewOutOfOrderCmdError("DATA command before MAIL command")
	} else if !s.IsAtLeastOneRcptReceived {
		return []byte{}, NewOutOfOrderCmdError("DATA command before at least one RCPT command")
	}
	message := "Start mail input; end with <CRLF>.<CRLF>"
	if err := s.Reply(StatusContinue, message); err != nil {
		return []byte{}, NewServerError(fmt.Sprintf("error sending hello %v", err))
	}
	body, err := s.receiveBody()
	if err != nil {
		return []byte{}, NewServerError(fmt.Sprintf("error reading body %v", err))
	}
	if err := s.Reply(StatusOk, "OK"); err != nil {
		return []byte{}, NewServerError(fmt.Sprintf("error sending hello %v", err))

	}
	return body, nil
}
func (s *Session) HandleQuit() error {
	message := fmt.Sprintf("%s service closing transmission channel", s.Server)
	if err := s.Reply(StatusClose, message); err != nil {
		return NewServerError(fmt.Sprintf("error sending hello %v", err))
	}
	return nil
}

func (s *Session) HandleUnknownError(err error) {
	message := fmt.Sprintf("unknown server error:%s", err.Error())
	if err := s.Reply(StatusUnknownError, message); err != nil {
		log.Printf("error sending hello %v", err)
	}
}
func (s *Session) Reply(statusCode int, statusLine string) error {
	if _, err := s.Conn.Write([]byte(fmt.Sprintf("%d %s\r\n", statusCode, statusLine))); err != nil {
		return err
	}
	return nil
}
func (s *Session) NextCMD() (Command, error) {
	buff := make([]byte, 1024)
	command := Command{}
	if _, err := s.Conn.Read(buff); err != nil {
		return Command{}, NewServerError(err.Error())
	}

	cmdIn := strings.Split(string(buff), "\r\n")
	if len(cmdIn) != 2 {
		return command, NewSyntaxError("invalid command format")
	}
	args := strings.Split(cmdIn[0], " ")
	command.Name = args[0]
	command.Args = args[1:]
	if command.Name == "MAIL" {
		err := command.ParseFrom()
		if err != nil {
			return command, NewSyntaxError("invalid command format: " + err.Error())
		}
	} else if command.Name == "RCPT" {
		err := command.ParseTo()
		if err != nil {
			return command, NewSyntaxError("invalid command format: " + err.Error())
		}
	}
	return command, nil
}

func (s *Session) GetMail() (storage.Envelope, error) {
	mail := storage.Envelope{}
	// handling smtp commands
	for {
		cmd, err := s.NextCMD()
		if err != nil {
			if errors.As(err, &SyntaxError{}) {
				if err = s.Reply(StatusSyntaxError, err.Error()); err != nil {
					log.Printf("wrong syntax %v", err)
					return storage.Envelope{}, err
				}
			} else if errors.As(err, &OutOfOrderCmdError{}) {
				if err = s.Reply(StatusOutOfSequenceCmdError, err.Error()); err != nil {
					log.Printf("out of sequence commands %v", err)
					return storage.Envelope{}, err
				}
			} else if errors.As(err, &ServerError{}) {
				return storage.Envelope{}, err
			}
			panic(fmt.Sprintf("error cmd %v", err))
		}
		switch cmd.Name {
		case "QUIT":
			err := s.HandleQuit()
			if err != nil {
				return mail, err
			}
			return mail, nil
		case "EHLO", "HELLO":
			err := s.HandleHello(cmd)
			if err != nil {
				return mail, err
			}
		case "MAIL":
			mail.Sender, err = s.HandleMail(cmd)
			if err != nil {
				return mail, err
			}
		case "RCPT":
			recipient, err := s.HandleRcpt(cmd)
			if err != nil {
				return mail, err
			}
			mail.Recipient = append(mail.Recipient, recipient)
		case "DATA":
			mail.Content, err = s.HandleData(cmd)
			if err != nil {
				return mail, err
			}
		}
	}
	return mail, nil
}
func (s *Session) receiveBody() ([]byte, error) {
	var message []byte
	//reads first 1GB or until encounter first \r\n.\r\n
	for i := 0; i < 1024*1024; i++ {
		buff := make([]byte, 1024)
		if _, err := s.Conn.Read(buff); err != nil {
			return nil, err
		}
		if strings.Contains(string(buff), "\r\n.\r\n") {
			line := strings.Split(string(buff), "\r\n.\r\n")
			message = append(message, line[0]...)
			return message, nil
		}
		message = append(message, buff...)
	}
	return message, NewServerError("large DATA, send less then 1 GB")
}
