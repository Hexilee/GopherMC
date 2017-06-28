package main

import (
	"net"
)

type TCPHub struct {
	Listener   *TCPHubListener
	Conn       *net.TCPConn
	Clients    map[*TCPClient]bool
	Register   chan *TCPClient
	Unregister chan *TCPClient
	Broadcast  chan []byte
	Receiver   chan []byte
	Signal     chan string
	Name       string
}


func (s *TCPHub) Start (conn *net.TCPConn, MaxBytes int) {
	go s.HandConn(conn, MaxBytes)
	go s.RegisterClient()
	go s.SendMessage()
	go s.ClientWriter()
}


func (s *TCPHub) ClientWriter() {
	for {
		select {
		case data := <-s.Broadcast:
			for client, in := range s.Clients {
				if in {
					client.Message <- data
				}
			}
		}
	}
}

func (s *TCPHub) RegisterClient() {
	for {
		select {
		case client := <-s.Register:
			s.Clients[client] = true
		case client := <-s.Unregister:
			delete(s.Clients, client)
		}
	}
}

func (s *TCPHub) SendMessage() {
	for {
		select {
		case message := <-s.Receiver:
			s.Conn.Write(message)
		}
	}
}

func (s *TCPHub) HandConn(conn *net.TCPConn, bytes int) {
	s.Conn = conn
	//s.Conn.SetReadDeadline(time.Now().Add(10 * time.Minute))
	defer s.Conn.Close()
	for {
		var data = make([]byte, bytes, bytes)
		_, err := s.Conn.Read(data)
		CheckErr(err)
		s.Broadcast <- data
	}
}

func NewTCPHub() *TCPHub {
	return &TCPHub{
		Register:   make(chan *TCPClient, 1000),
		Unregister: make(chan *TCPClient, 1000),
		Broadcast:  make(chan []byte, 2000),
		Receiver:   make(chan []byte, 2000),
		Clients:    make(map[*TCPClient]bool),
		Signal:     make(chan string, 100),
	}
}
