package main

import (
	"context"
	"net"
)

type SocketHub struct {
	Service    *Service
	Listener   *SocketHubListener
	Conn       net.Conn
	Clients    map[*SocketClient]bool
	Register   chan *SocketClient
	Unregister chan *SocketClient
	Broadcast  chan []byte
	Receiver   chan []byte
	Signal     chan string
	Name       string
	Context    context.Context
	Cancel     context.CancelFunc
}

func (s *SocketHub) Start(conn net.Conn, MaxBytes int) {
	go s.HandConn(conn, MaxBytes)
	go s.RegisterClient()
	go s.SendMessage()
	go s.ClientWriter()
}

func (s *SocketHub) ClientWriter() {

	defer func() {
		p := recover()
		CheckPanic(p, s.Service, "Hub ClientWriter panic!")
	}()

Circle:
	for {
		select {
		case <-s.Context.Done():
			s.Service.Info <- "Socket Hub ClientWriter Done. Addr: " + s.Conn.RemoteAddr().String()
			break Circle
		case data := <-s.Broadcast:
			for client, in := range s.Clients {
				if in {
					client.Message <- data
				}
			}
		}
	}
}

func (s *SocketHub) RegisterClient() {

	defer func() {
		p := recover()
		CheckPanic(p, s.Service, "Hub RegisterClient panic!")
	}()

Circle:
	for {
		select {
		case <-s.Context.Done():
			s.Service.Info <- "Socket Hub RegisterClient Done. Addr: " + s.Conn.RemoteAddr().String()
			break Circle
		case client := <-s.Register:
			s.Clients[client] = true
		case client := <-s.Unregister:
			delete(s.Clients, client)
		}
	}
}

func (s *SocketHub) SendMessage() {

	defer func() {
		p := recover()
		CheckPanic(p, s.Service, "Hub SendMessage panic!")
	}()

Circle:
	for {
		select {
		case <-s.Context.Done():
			s.Service.Info <- "Socket Hub SendMessage Done. Addr: " + s.Conn.RemoteAddr().String()
			break Circle
		case message := <-s.Receiver:
			s.Conn.Write(message)
		}
	}
}

func (s *SocketHub) HandConn(conn net.Conn, bytes int) {

	defer func() {
		s.Conn.Close()
		s.Listener.Unregister <- s
		p := recover()
		CheckPanic(p, s.Service, "Hub HandConn panic!")
	}()

	s.Conn = conn
	//s.Conn.SetReadDeadline(time.Now().Add(10 * time.Minute))
	defer s.Conn.Close()
	for {
		var data = make([]byte, bytes, bytes)
		_, err := s.Conn.Read(data)
		if !DealConnErr(err, conn, s.Service) {
			s.Service.Info <- "Socket Hub HandConn Done. Addr: " + s.Conn.RemoteAddr().String()
			s.Cancel()
			break
		}
		s.Broadcast <- data
	}
}

func (s *SocketHub) Clean() (ok bool) {

	defer func() {
		p := recover()
		if !CheckPanic(p, s.Service, "Hub Clean panic!") {
			ok = false
		}
	}()

	close(s.Register)
	close(s.Unregister)
	close(s.Receiver)
	close(s.Signal)
	close(s.Broadcast)
	s.Register = make(chan *SocketClient, 1000)
	s.Unregister = make(chan *SocketClient, 1000)
	s.Broadcast = make(chan []byte, 2000)
	s.Receiver = make(chan []byte, 2000)
	s.Signal = make(chan string, 100)
	s.Clients = make(map[*SocketClient]bool)
	ok = true
	return
}

func NewSocketHub() *SocketHub {
	return &SocketHub{
		Register:   make(chan *SocketClient, 1000),
		Unregister: make(chan *SocketClient, 1000),
		Broadcast:  make(chan []byte, 2000),
		Receiver:   make(chan []byte, 2000),
		Clients:    make(map[*SocketClient]bool),
		Signal:     make(chan string, 100),
		//Context:    context.Background(),
		//Cancel:     func() {},
	}
}
