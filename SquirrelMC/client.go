package main

import (
	"net"
	"time"
)

type TCPClient struct {
	Conn    *net.TCPConn
	Hub     *TCPHub
	Message chan []byte
}

func (s *TCPClient) HandConn(conn *net.TCPConn, bytes int,  hub *TCPHub) {
	s.Conn = conn
	s.Hub = hub
	s.Hub.Register <- s
	s.Conn.SetReadDeadline(time.Now().Add(10 * time.Minute))
	defer s.Conn.Close()
	for {
		var data = make([]byte, bytes, bytes)
		_, err := s.Conn.Read(data)
		CheckErr(err)
		s.Hub.Receiver <- data
		s.Conn.Write([]byte("Received"))
	}
}

func (s *TCPClient) Broadcast() {
	for {
		select {
		case message := <- s.Message:
			s.Conn.Write(message)
		}
	}
}