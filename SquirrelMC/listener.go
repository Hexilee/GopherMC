package main

import (
	"net"
	"log"
	S "golang.org/x/sync/syncmap"
)

type TCPHubListener struct {
	Listener   *net.TCPListener
	Signal     chan string
	HubTable   *S.Map
	Unregister chan *TCPHub
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
	log.Printf("Hub listener listening at %s", service)

	return &TCPHubListener{
		Listener: listener,
		Signal:   make(chan string, 1000),
		HubTable: &S.Map{},
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
		case signal := <- t.Signal:
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

	go newHub.Start(conn, MaxBytes)
}
