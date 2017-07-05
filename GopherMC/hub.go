package main

import (
	"context"
	"net"
	"encoding/binary"
	"errors"
	"bytes"
	"bufio"
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

func (s *SocketHub) Start(conn net.Conn) {
	go s.HandConn(conn)
	go s.RegisterClient()
	go s.SendMessage()
	go s.ClientWriter()
}

func (s *SocketHub) ClientWriter() {

	defer func() {
		p := recover()
		CheckPanic(p,"Hub ClientWriter panic!")
	}()

Circle:
	for {
		select {
		case <-s.Context.Done():
			Logger.Info("Socket Hub ClientWriter Done. Addr: " + s.Conn.RemoteAddr().String())
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
		CheckPanic(p,"Hub RegisterClient panic!")
	}()

Circle:
	for {
		select {
		case <-s.Context.Done():
			Logger.Info("Socket Hub RegisterClient Done. Addr: " + s.Conn.RemoteAddr().String())
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
		CheckPanic(p,"Hub SendMessage panic!")
	}()

Circle:
	for {
		select {
		case <-s.Context.Done():
			Logger.Info("Socket Hub SendMessage Done. Addr: " + s.Conn.RemoteAddr().String())
			break Circle
		case message := <-s.Receiver:
			SecureWrite(message, s.Conn)
		}
	}
}

func (s *SocketHub) split(data []byte, atEOF bool) (adv int, token []byte, err error) {
	length := len(data)
	if length < headerLen {
		return 0, nil, nil
	}
	if length > 1048576 { //1024*1024=1048576
		Logger.Info("Socket Hub %s Read Error. Addr: %s", s.Name, s.Conn.RemoteAddr().String())
		s.Cancel()
		return 0, nil, errors.New("too large data!")
	}
	var lhead uint32
	buf := bytes.NewReader(data)
	binary.Read(buf, binary.LittleEndian, &lhead)

	tail := length - headerLen
	if lhead > 1048576 {
		Logger.Info("Socket Hub %s Read Error. Addr: %s", s.Name, s.Conn.RemoteAddr().String())
		s.Cancel()
		return 0, nil, errors.New("too large data!")
	}
	if uint32(tail) < lhead {
		return 0, nil, nil
	}
	adv = headerLen + int(lhead)
	token = data[:adv]
	return adv, token, nil
}

func (s *SocketHub) Scan() {
	scanner := bufio.NewScanner(s.Conn)
	scanner.Split(s.split)

Circle:
	for scanner.Scan() {
		select {
		case <-s.Context.Done():
			Logger.Info("Socket Hub %s HandConn Done. Addr: %s", s.Name, s.Conn.RemoteAddr().String())
			break Circle
		default:
		}

		data := scanner.Bytes()
		msg := make([]byte, len(data))
		copy(msg, data)
		s.Broadcast <- msg
	}
	if scanner.Err() != nil {
		err := scanner.Err()
		Logger.Error(err.Error())
	}
}


func (s *SocketHub) HandConn(conn net.Conn) {

	defer func() {
		s.Conn.Close()
		s.Cancel()
		s.Listener.Unregister <- s
		p := recover()
		CheckPanic(p, "Hub HandConn panic!")
	}()

	s.Conn = conn
	s.Scan()
}

func (s *SocketHub) Clean() (ok bool) {

	defer func() {
		p := recover()
		if !CheckPanic(p,"Hub Clean panic!") {
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
	}
}
