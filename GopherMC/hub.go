package main

import (
	"io"
)

type SocketHub struct {
	Service    *Service
	Listener   *SocketHubListener
	Conn       io.ReadWriteCloser
	Clients    map[*SocketClient]bool
	Register   chan *SocketClient
	Unregister chan *SocketClient
	Broadcast  chan []byte
	Receiver   chan []byte
	Signal     chan string
	Name       string
}

func (s *SocketHub) Start(conn io.ReadWriteCloser, MaxBytes int) {
	go s.HandConn(conn, MaxBytes)
	go s.RegisterClient()
	go s.SendMessage()
	go s.ClientWriter()
}

func (s *SocketHub) ClientWriter() {
Circle:
	for {
		select {
		case data := <-s.Broadcast:
			for client, in := range s.Clients {
				if in {
					client.Message <- data
				}
			}
		case signal := <-s.Signal:
			if signal == "kill" {
				s.Signal <- "kill"
				for client, in := range s.Clients {
					if in {
						client.Conn.Close()
						delete(s.Clients, client)
						client.Signal <- "kill"
					}
				}
				break Circle
			}
		}
	}
}

func (s *SocketHub) RegisterClient() {
Circle:
	for {
		select {
		case client := <-s.Register:
			s.Clients[client] = true
		case client := <-s.Unregister:
			delete(s.Clients, client)
		case signal := <-s.Signal:
			if signal == "kill" {
				s.Signal <- "kill"
				break Circle
			}
		}
	}
}

func (s *SocketHub) SendMessage() {
Circle:
	for {
		select {
		case message := <-s.Receiver:
			s.Conn.Write(message)
		case signal := <-s.Signal:
			if signal == "kill" {
				s.Signal <- "kill"
				break Circle
			}
		}
	}
}

func (s *SocketHub) HandConn(conn io.ReadWriteCloser, bytes int) {
	s.Conn = conn
	//s.Conn.SetReadDeadline(time.Now().Add(10 * time.Minute))
	defer s.Conn.Close()
	for {
		var data = make([]byte, bytes, bytes)
		_, err := s.Conn.Read(data)
		if DealConnErr(err, conn) {
			s.Listener.Unregister <- s
		}
		s.Broadcast <- data
	}
}

func NewSocketHub() *SocketHub {
	return &SocketHub{
		Register:   make(chan *SocketClient, 1000),
		Unregister: make(chan *SocketClient, 1000),
		Broadcast:  make(chan []byte, 2000),
		Receiver:   make(chan []byte, 2000),
		Clients:    make(map[*SocketClient]bool),
		Signal:     make(chan string, 100),
	}
}
