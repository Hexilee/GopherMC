package main

import (
	"io"
	"context"
)

type SocketClient struct {
	Service  *Service
	Conn     io.ReadWriteCloser
	Listener *SocketClientListener
	Hub      *SocketHub
	Message  chan []byte
	Signal   chan string
	Context  context.Context
	Cancel   context.CancelFunc
}

func (s *SocketClient) HandConn(conn io.ReadWriteCloser, bytes int) {

	defer func() {
		s.Conn.Close()
		p := recover()
		CheckPanic(p, s.Service, "Client HandConn panic")
	}()

	s.Conn = conn
	s.Hub.Register <- s
	//s.Conn.SetReadDeadline(time.Now().Add(10 * time.Minute))
	//SocketRead(s.Conn, s.Hub.Receiver, s.Service)
	for {
		var data = make([]byte, bytes, bytes)
		//_, err := s.Conn.Read(data)
		//CheckErr(err)
		SecureRead(data, conn, s.Service)
		s.Hub.Receiver <- data
	}
}

func (s *SocketClient) Broadcast() {
Circle:
	for {
		select {
		case message := <-s.Message:
			s.Conn.Write(message)
		case signal := <-s.Signal:
			if signal == "kill" {
				s.Conn.Close()
				break Circle
			}
		}
	}
	defer func() {
		s.Conn.Close()
		p := recover()
		CheckPanic(p, s.Service, "Client Broadcast panic")
	}()
}

func NewSocketClient() *SocketClient {
	return &SocketClient{
		Message: make(chan []byte, 100),
		Signal:  make(chan string, 100),
	}
}
