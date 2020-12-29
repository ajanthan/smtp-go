package smtp

import (
	"errors"
	"fmt"
	"github/ajanthan/smtp-go/pkg/storage"
	"log"
	"net"
	"net/mail"
	"net/textproto"
	"strings"
)

type Session struct {
	conn                     *net.Conn
	Conn                     *textproto.Conn
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
func (s *Session) HandleData(Command) (*mail.Message, error) {
	if !s.IsHelloReceived {
		return nil, NewOutOfOrderCmdError("DATA command before EHLO/HELLO command")
	} else if !s.IsMailReceived {
		return nil, NewOutOfOrderCmdError("DATA command before MAIL command")
	} else if !s.IsAtLeastOneRcptReceived {
		return nil, NewOutOfOrderCmdError("DATA command before at least one RCPT command")
	}
	message := "Start mail input; end with <CRLF>.<CRLF>"
	if err := s.Reply(StatusContinue, message); err != nil {
		return nil, NewServerError(fmt.Sprintf("error sending hello %v", err))
	}
	msg, err := mail.ReadMessage(s.Conn.DotReader())
	if err != nil {
		return nil, NewServerError(fmt.Sprintf("error reading mail body %v", err))
	}
	if err := s.Reply(StatusOk, "OK"); err != nil {
		return nil, NewServerError(fmt.Sprintf("error sending ok %v", err))
	}
	return msg, nil
}
func (s *Session) HandleQuit() error {
	message := fmt.Sprintf("%s service closing transmission channel", s.Server)
	if err := s.Reply(StatusClose, message); err != nil {
		return NewServerError(fmt.Sprintf("error sending hello %v", err))
	}
	_ = s.Conn.Close()
	return nil
}

func (s *Session) HandleReset() error {
	s.IsMailReceived = false
	s.IsAtLeastOneRcptReceived = false
	if err := s.Reply(StatusOk, "OK"); err != nil {
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
	if err := s.Conn.PrintfLine("%d %s", statusCode, statusLine); err != nil {
		return err
	}
	return nil
}
func (s *Session) NextCMD() (Command, error) {
	command := Command{}
	buff, err := s.Conn.ReadLine()
	if err != nil {
		return Command{}, NewSyntaxError(err.Error())
	}
	args := strings.Split(buff, " ")
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
	envelope := storage.Envelope{}
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
			panic(fmt.Sprintf("error cmd %s", err))
		}
		switch cmd.Name {
		case "QUIT":
			err := s.HandleQuit()
			if err != nil {
				return envelope, err
			}
			return envelope, nil
		case "EHLO", "HELO":
			err := s.HandleHello(cmd)
			if err != nil {
				return envelope, err
			}
		case "MAIL":
			envelope.Sender, err = s.HandleMail(cmd)
			if err != nil {
				return envelope, err
			}
		case "RCPT":
			recipient, err := s.HandleRcpt(cmd)
			if err != nil {
				return envelope, err
			}
			envelope.Recipient = append(envelope.Recipient, recipient)
		case "DATA":
			envelope.Content, err = s.HandleData(cmd)
			if err != nil {
				return envelope, err
			}
		case "RSET":
			envelope = storage.Envelope{}
			err := s.HandleReset()
			if err != nil {
				return envelope, err
			}
		}
	}
	return envelope, nil
}
