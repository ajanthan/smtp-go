package smtp

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github/ajanthan/smtp-go/pkg/storage"
	"log"
	"net"
	"net/mail"
	"net/textproto"
	"strings"
)

const (
	StartTLS = "STARTTLS"
	Auth     = "AUTH"
)

//TODO: support AUTH
type Session struct {
	conn                     *net.Conn
	Conn                     *textproto.Conn
	Server                   string
	Client                   string
	IsHelloReceived          bool
	IsMailReceived           bool
	IsAtLeastOneRcptReceived bool
	IsAuthenticated          bool
	IsTLSConn                bool
	TLSConfig                *tls.Config
	Extensions               []string
	Auth                     *AuthenticationService
	Secure                   bool
}

func (s *Session) Start() error {
	greetings := s.Server + " ESMTP smtp-go"
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
	if len(s.Extensions) == 0 {
		if err := s.Reply(StatusOk, message); err != nil {
			return NewServerError(fmt.Sprintf("error sending ok %v", err))
		}
	} else {
		if err := s.MultiReply(StatusOk, message); err != nil {
			return NewServerError(fmt.Sprintf("error sending ok %v", err))
		}
		for i, extension := range s.Extensions {
			if i != len(s.Extensions)-1 {
				if err := s.MultiReply(StatusOk, extension); err != nil {
					return NewServerError(fmt.Sprintf("error sending ok %v", err))
				}
			}
		}
		if err := s.Reply(StatusOk, s.Extensions[len(s.Extensions)-1]); err != nil {
			return NewServerError(fmt.Sprintf("error sending ok %v", err))
		}
	}

	s.IsHelloReceived = true
	return nil
}
func (s *Session) HandleMail(cmd Command) (string, error) {
	err := s.checkAuthRequired()
	if err != nil {
		return "", err
	}

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
	err := s.checkAuthRequired()
	if err != nil {
		return "", err
	}
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
	err := s.checkAuthRequired()
	if err != nil {
		return nil, err
	}
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

func (s *Session) HandleStartTLS() error {
	if err := s.Reply(StatusReady, "Go ahead"); err != nil {
		return NewServerError(fmt.Sprintf("error sending hello %v", err))
	}
	s.IsHelloReceived = false
	s.IsAtLeastOneRcptReceived = false
	s.IsMailReceived = false
	tlsConn := tls.Server(*s.conn, s.TLSConfig)
	s.Conn = textproto.NewConn(tlsConn)
	s.IsTLSConn = true
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
func (s *Session) MultiReply(statusCode int, statusLine string) error {
	if err := s.Conn.PrintfLine("%d-%s", statusCode, statusLine); err != nil {
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

func (s *Session) GetMail() (*storage.Envelope, error) {
	envelope := storage.NewEnvelope(s.Server)
	// handling smtp commands
	for {
		cmd, err := s.NextCMD()
		if err != nil {
			if errors.As(err, &SyntaxError{}) {
				if err = s.Reply(StatusSyntaxError, err.Error()); err != nil {
					log.Printf("wrong syntax %v", err)
					return nil, err
				}
			} else if errors.As(err, &OutOfOrderCmdError{}) {
				if err = s.Reply(StatusOutOfSequenceCmdError, err.Error()); err != nil {
					log.Printf("out of sequence commands %v", err)
					return nil, err
				}
			} else if errors.As(err, &ServerError{}) {
				return nil, err
			}
			panic(fmt.Sprintf("error cmd %s", err))
		}
		switch cmd.Name {
		case "QUIT":
			err := s.HandleQuit()
			if err != nil {
				return nil, err
			}
			return envelope, nil
		case "EHLO", "HELO":
			err := s.HandleHello(cmd)
			if err != nil {
				return nil, err
			}
		case "MAIL":
			envelope.Sender, err = s.HandleMail(cmd)
			if err != nil {
				return nil, err
			}
		case "RCPT":
			recipient, err := s.HandleRcpt(cmd)
			if err != nil {
				return nil, err
			}
			envelope.Recipient = append(envelope.Recipient, recipient)
		case "DATA":
			envelope.Content, err = s.HandleData(cmd)
			if err != nil {
				return nil, err
			}
		case "RSET":
			envelope = storage.NewEnvelope(s.Server)
			err := s.HandleReset()
			if err != nil {
				return nil, err
			}
		case "STARTTLS":
			if s.TLSConfig != nil {
				envelope = storage.NewEnvelope(s.Server)
				err := s.HandleStartTLS()
				if err != nil {
					return nil, err
				}
			} else {
				if err := s.Reply(StatusCommandNotImplemented, fmt.Sprintf("%s is not supported", cmd.Name)); err != nil {
					return nil, NewServerError(fmt.Sprintf("error sending reply %v", err))
				}
			}
		case "AUTH":
			if s.Auth != nil {
				err := s.HandleAuth(cmd.Args, envelope.MessageID)
				if err != nil {
					return nil, err
				}
			} else {
				if err := s.Reply(StatusCommandNotImplemented, fmt.Sprintf("%s is not supported", cmd.Name)); err != nil {
					return nil, NewServerError(fmt.Sprintf("error sending reply %v", err))
				}
			}
		default:
			if err := s.Reply(StatusCommandNotImplemented, fmt.Sprintf("%s is not supported", cmd.Name)); err != nil {
				return nil, NewServerError(fmt.Sprintf("error sending reply %v", err))
			}
		}
	}
	return envelope, nil
}

func (s *Session) HandleAuth(args []string, messageID string) error {
	if !s.IsAuthenticated {
		switch args[0] {
		case "PLAIN":
			cred := ""
			if !s.IsTLSConn {
				if err := s.Reply(StatusTLSRequired, "TLS required for the AUTH command"); err != nil {
					return NewServerError(fmt.Sprintf("error sending reply %v", err))
				}
				return NewServerError("TLS required for the AUTH command")
			} else if len(args) > 1 {
				cred = args[1]
			} else {
				if err := s.Reply(StatusAuthChallenge, ""); err != nil {
					return NewServerError(fmt.Sprintf("error sending auth challenge %v", err))
				}
				cmd, err := s.NextCMD()
				if err != nil {
					return NewServerError(fmt.Sprintf("error receiving PLAIN credential %v", err))
				}
				cred = cmd.Name
			}
			if err := HandlePlainAuth(cred, *s.Auth); err != nil {
				return s.handleAuthError(err)
			}
			if err := s.Reply(StatusAuthSuccess, "Authentication successful"); err != nil {
				return NewServerError(fmt.Sprintf("error sending auth challenge %v", err))
			}

		case "LOGIN":
			if !s.IsTLSConn {
				if err := s.Reply(StatusTLSRequired, "TLS required for the AUTH command"); err != nil {
					return NewServerError(fmt.Sprintf("error sending reply %v", err))
				}
				return NewServerError("TLS required for the AUTH command")
			}
			username, err := s.getLoginParameters("VXNlcm5hbWU6")
			if err != nil {
				return err
			}
			password, err := s.getLoginParameters("UGFzc3dvcmQ6")
			if err != nil {
				return err
			}
			if err := HandleLoginAuth(username, password, *s.Auth); err != nil {
				return s.handleAuthError(err)
			}
			if err := s.Reply(StatusAuthSuccess, "Authentication successful"); err != nil {
				return NewServerError(fmt.Sprintf("error sending auth challenge %v", err))
			}

		case "CRAM-MD5":
			challenge := base64Encode(messageID)
			if err := s.Reply(StatusAuthChallenge, string(challenge)); err != nil {
				return NewServerError(fmt.Sprintf("error sending auth challenge %v", err))
			}
			cmd, err := s.NextCMD()
			if err != nil {
				return NewServerError(fmt.Sprintf("error receiving PLAIN credential %v", err))
			}
			if err := HandleMD5CRAMAuth(cmd.Name, []byte(messageID), *s.Auth); err != nil {
				return s.handleAuthError(err)
			}
			if err := s.Reply(StatusAuthSuccess, "Authentication successful"); err != nil {
				return NewServerError(fmt.Sprintf("error sending auth challenge %v", err))
			}
		}
	} else {
		if err := s.Reply(StatusOutOfSequenceCmdError, "AUTH is already done"); err != nil {
			return NewServerError(fmt.Sprintf("error sending reply %v", err))
		}
	}
	return nil
}

func (s *Session) getLoginParameters(msg string) (string, error) {
	if err := s.Reply(StatusAuthChallenge, msg); err != nil {
		return "", NewServerError(fmt.Sprintf("error sending auth challenge %v", err))
	}
	cmd, err := s.NextCMD()
	if err != nil {
		return "", NewServerError(fmt.Sprintf("error receiving Login credential %v", err))
	}
	return cmd.Name, nil
}
func (s *Session) handleAuthError(err error) error {
	if errors.As(err, &InvalidCredentialError{}) {
		if err = s.Reply(StatusInvalidCredentialError, err.Error()); err != nil {
			return NewServerError(fmt.Sprintf("error sending reply %v", err))
		}
	} else if errors.As(err, &ServerError{}) {
		if err = s.Reply(StatusTempAuthError, err.Error()); err != nil {
			return NewServerError(fmt.Sprintf("error sending reply %v", err))
		}
	}
	return err

}
func (s *Session) checkAuthRequired() error {
	if s.Secure && !s.IsAuthenticated {
		if err := s.Reply(StatusAuthRequired, "Authentication required"); err != nil {
			return NewServerError(fmt.Sprintf("error sending reply %v", err))
		}
		return NewAuthRequiredError("Authentication required")
	}
	return nil
}
