package main

import (
	"context"
	"net"
	"encoding/binary"
	"bufio"
	"errors"
	"bytes"
)

const (
	headerLen int = 4
)

type SocketClient struct {
	Service  *Service
	Conn     net.Conn
	Listener *SocketClientListener
	Hub      *SocketHub
	Message  chan []byte
	Signal   chan string
	Context  context.Context
	Cancel   context.CancelFunc
}

func (s *SocketClient) HandConn(conn net.Conn, bytes int) {

	defer func() {
		s.Conn.Close()
		if s.Clean() {
			s.Listener.ClientRecycler <- s
		}
		p := recover()
		CheckPanic(p, s.Service, "Client HandConn panic")
	}()

	s.Conn = conn
	s.Hub.Register <- s
	s.Scan()
}

func (s *SocketClient) Broadcast() {
Circle:
	for {
		select {
		case <-s.Context.Done():
			s.Service.Info <- "Socket Client Broadcast Done. Addr: " + s.Conn.RemoteAddr().String()
			break Circle
		case message := <-s.Message:
			if !SecureWrite(message, s.Conn, s.Service) {
				s.Cancel()
			}
		}
	}
}

func (s *SocketClient) Clean() (ok bool) {
	defer func() {
		p := recover()
		if !CheckPanic(p, s.Service, "Hub Clean panic!") {
			ok = false
		}
	}()

	close(s.Message)
	close(s.Signal)
	s.Message = make(chan []byte, 100)
	s.Signal = make(chan string, 5)

	ok = true
	return
}

func (s *SocketClient) split(data []byte, atEOF bool) (adv int, token []byte, err error) {
	length := len(data)
	if length < headerLen {
		return 0, nil, nil
	}
	if length > 1048576 { //1024*1024=1048576
		s.Service.Info <- "Socket Client Read Error. Addr: " + s.Conn.RemoteAddr().String()
		s.Cancel()
		return 0, nil, errors.New("too large data!")
	}
	var lhead uint32
	buf := bytes.NewReader(data)
	binary.Read(buf, binary.LittleEndian, &lhead)

	tail := length - headerLen
	if lhead > 1048576 {
		s.Service.Info <- "Socket Client Read Error. Addr: " + s.Conn.RemoteAddr().String()
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

func (s *SocketClient) Scan() {
	scanner := bufio.NewScanner(s.Conn)
	scanner.Split(s.split)

Circle:
	for scanner.Scan() {
		select {
		case <-s.Context.Done():
			s.Service.Info <- "Socket Client HandConn Done. Addr: " + s.Conn.RemoteAddr().String()
			break Circle
		default:
		}

		data := scanner.Bytes()
		msg := make([]byte, len(data))
		copy(msg, data)
		s.Hub.Receiver <- msg
	}
	if scanner.Err() != nil {
		err := scanner.Err()
		s.Service.Error <- &err
	}
}



func NewSocketClient() *SocketClient {
	return &SocketClient{
		Message: make(chan []byte, 100),
		Signal:  make(chan string, 5),
		//Context: context.Background(),
		//Cancel:  func() {},
	}
}
