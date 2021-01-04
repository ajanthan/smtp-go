package smtp

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/textproto"
	"strconv"
)

type Server struct {
	Address     string
	SMTPPort    int
	Receiver    MailReceiver
	TLSConfig   *tls.Config
	AuthService AuthenticationService
	Secure      bool
	ConnTimeOut int
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
				Conn:        textproto.NewConn(conn),
				conn:        conn,
				Server:      s.Address,
				Secure:      s.Secure,
				Auth:        s.AuthService,
				Receiver:    s.Receiver,
				ConnTimeOut: s.ConnTimeOut,
			}
			if s.TLSConfig != nil {
				session.TLSConfig = s.TLSConfig
				session.Extensions = append(session.Extensions, StartTLS)
			}
			if s.Secure {
				session.Extensions = append(session.Extensions, Auth)
			}
			err := session.Handle()
			if err != nil {
				session.HandleUnknownError(err)
				log.Printf("error starting session %v", err)
			}
		}()
	}
}
