package smtp

import (
	"crypto/tls"
	"fmt"
	"github/ajanthan/smtp-go/pkg/storage"
	"log"
	"net"
	"net/textproto"
	"strconv"
)

type Server struct {
	Address   string
	SMTPPort  int
	Receiver  MailReceiver
	Storage   storage.Storage
	TLSConfig *tls.Config
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
		go func() {
			session := &Session{
				Conn:   textproto.NewConn(conn),
				conn:   &conn,
				Server: s.Address,
			}
			if s.TLSConfig != nil {
				session.TLSConfig = s.TLSConfig
			}
			err := session.Start()
			if err != nil {
				session.HandleUnknownError(err)
				log.Printf("error starting session %v", err)
			}
			mail, err := session.GetMail()
			if err != nil {
				session.HandleUnknownError(err)
				log.Printf("error Receiving mail %v", err)
			} else {
				err := s.Receiver.Receive(mail)
				if err != nil {
					log.Printf("error handling received mail %v", err)
				}
			}
		}()
	}
}
