package main

import (
	"net"
	S "golang.org/x/sync/syncmap"
)

type TCPHubListener struct {
	Service    *Service
	Listener   *net.TCPListener
	Signal     chan string
	HubTable   *S.Map
	Unregister chan *TCPHub
}

type TCPClientListener struct {
	Service  *Service
	Listener *net.TCPListener
	Signal   chan string
}

func NewTCPHubListener(service string) *TCPHubListener {

	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	if !CheckErr(err) {
		return nil
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if !CheckErr(err) {
		return nil
	}

	logger.Info("Hub listener listening at %s", service)

	return &TCPHubListener{
		Listener:   listener,
		Signal:     make(chan string, 1000),
		HubTable:   &S.Map{},
		Unregister: make(chan *TCPHub, 100),
	}
}

func NewTCPClientListener(service string) *TCPClientListener {

	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	if !CheckErr(err) {
		return nil
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if !CheckErr(err) {
		return nil
	}
	logger.Info("Client listener listening at %s", service)

	return &TCPClientListener{
		Listener: listener,
		Signal:   make(chan string, 1000),
	}
}

func (t *TCPHubListener) Start(MaxBytes int) {
	for {
		conn, err := t.Listener.AcceptTCP()
		if !CheckErr(err) {
			continue
		}
		if !SecureWrite([]byte("Fine, connected\n"), conn) {
			continue
		}

		go t.HandConn(conn, MaxBytes)
	}
}

func (t *TCPHubListener) RemoveHub() {
Circle:
	for {
		select {
		case deadHub := <-t.Unregister:
			deadHub.Conn.Close()
			deadHub.Signal <- "kill"
			t.HubTable.Delete(deadHub.Name)
		case signal := <-t.Signal:
			if signal == "kill" {
				break Circle
			}
		}
	}
}

func (t *TCPHubListener) HandConn(conn *net.TCPConn, MaxBytes int) {
	var registerInfo [32]byte
	conn.Read(registerInfo[:])
	connName := string(registerInfo[:])

	newHub := NewTCPHub()
	_, ok := t.HubTable.LoadOrStore(connName, newHub)
	if ok {
		conn.Write([]byte("The hub already exist!\n"))
		conn.Close()
		return
	}
	newHub.Conn = conn
	if !SecureWrite([]byte("Register successfully!\n"), conn) {
		return
	}
	newHub.Listener = t
	newHub.Name = connName
	newHub.Service = t.Service

	go newHub.Start(conn, MaxBytes)
}

func (t *TCPClientListener) Start(MaxBytes int, HubTable *S.Map) {
	for {
		conn, err := t.Listener.AcceptTCP()
		if !CheckErr(err) {
			continue
		}
		if !SecureWrite([]byte("Fine, connected\n"), conn) {
			continue
		}

		go t.HandConn(conn, MaxBytes, HubTable)
	}
}

func (t *TCPClientListener) HandConn(conn *net.TCPConn, MaxBytes int, HubTable *S.Map) {
	var registerInfo [32]byte
	conn.Read(registerInfo[:])
	connName := string(registerInfo[:])
	hub, ok := HubTable.Load(connName)
	if !ok {
		conn.Write([]byte("No hub named " + connName))
		conn.Close()
		return
	}
	actualHub, ok := hub.(*TCPHub)
	if !ok {
		conn.Write([]byte("Bad Gate\n"))
		conn.Close()
		return
	}

	if !SecureWrite([]byte("Register successfully!\n"), conn) {
		return
	}
	newClient := NewTCPClient()
	newClient.Service = t.Service
	go newClient.HandConn(conn, MaxBytes, actualHub)
	go newClient.Broadcast()
}
