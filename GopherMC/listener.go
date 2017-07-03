package main

import (
	"net"
	S "golang.org/x/sync/syncmap"
	"fmt"
	"context"
)

type SocketHubListener struct {
	Service     *Service
	Listener    *net.TCPListener
	Signal      chan string
	HubTable    *S.Map
	Unregister  chan *SocketHub
	HubRecycler chan *SocketHub
	Context     context.Context
	Cancel      context.CancelFunc
}

type SocketClientListener struct {
	Service        *Service
	Listener       *net.TCPListener
	Signal         chan string
	ClientRecycler chan *SocketClient
	Context        context.Context
	Cancel         context.CancelFunc
}

func NewSocketHubListener(service string, srv *Service) *SocketHubListener {

	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	if !CheckErr(err, srv) {
		return nil
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if !CheckErr(err, srv) {
		return nil
	}

	//logger.Info("Hub listener listening at %s", service)
	srv.Info <- fmt.Sprintf("Hub listener listening at %s", service)

	return &SocketHubListener{
		Listener:    listener,
		Signal:      make(chan string, 5),
		HubTable:    &S.Map{},
		Unregister:  make(chan *SocketHub, 1000),
		HubRecycler: make(chan *SocketHub, 1000),
		//Context:     context.Background(),
		//Cancel:      func() {},
	}
}

func NewSocketClientListener(service string, srv *Service) *SocketClientListener {

	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	if !CheckErr(err, srv) {
		return nil
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if !CheckErr(err, srv) {
		return nil
	}
	//logger.Info("Client listener listening at %s", service)
	srv.Info <- fmt.Sprintf("Client listener listening at %s", service)

	return &SocketClientListener{
		Listener:       listener,
		Signal:         make(chan string, 5),
		ClientRecycler: make(chan *SocketClient, 10000),
		//Context:        context.Background(),
		//Cancel:         func() {},
	}
}

func (t *SocketHubListener) Start(srv *Service) {
	for {
		conn, err := t.Listener.AcceptTCP()
		if !CheckErr(err, srv) {
			continue
		}
		if !SecureWrite([]byte("Fine, connected\n"), conn, srv) {
			continue
		}

		go t.HandConn(conn, srv)
	}
}

func (t *SocketHubListener) RecycleHub() {
Circle:
	for {
		select {
		case <-t.Context.Done():
			break Circle
		case deadHub := <-t.Unregister:
			if deadHub.Clean() {
				t.HubRecycler <- deadHub
			}
			t.HubTable.Delete(deadHub.Name)
		}
	}
}

func (t *SocketHubListener) HandConn(conn *net.TCPConn, srv *Service) {
	var registerInfo [32]byte
	conn.Read(registerInfo[:])
	connName := string(registerInfo[:])

	_, ok := t.HubTable.Load(connName)
	if ok {
		conn.Write([]byte("The hub already exist!\n"))
		conn.Close()
		return
	}

	var newHub *SocketHub
	select {
	case binHub := <-t.HubRecycler:
		newHub = binHub
		newHub.Name = connName
	default:
		newHub = NewSocketHub()
		newHub.Listener = t
		newHub.Name = connName
		newHub.Service = t.Service
	}

	t.HubTable.Store(connName, newHub)

	socketHubCtx, socketHubCancel := context.WithCancel(t.Context)
	newHub.Context = socketHubCtx
	newHub.Cancel = socketHubCancel
	newHub.Conn = conn
	if !SecureWrite([]byte("Register successfully!\n"), conn, srv) {
		return
	}
	t.Service.Info <- conn.RemoteAddr().String() + " hub register as " + connName
	go newHub.Start(conn)
}

func (t *SocketClientListener) Start(HubTable *S.Map, srv *Service) {
	for {
		conn, err := t.Listener.AcceptTCP()
		if !CheckErr(err, srv) {
			continue
		}
		if !SecureWrite([]byte("Fine, connected\n"), conn, srv) {
			continue
		}

		go t.HandConn(conn, HubTable, srv)
	}
}

func (t *SocketClientListener) HandConn(conn *net.TCPConn, HubTable *S.Map, srv *Service) {
	var registerInfo [32]byte
	conn.Read(registerInfo[:])
	connName := string(registerInfo[:])
	hub, ok := HubTable.Load(connName)
	if !ok {
		conn.Write([]byte("No hub named " + connName))
		conn.Close()
		return
	}
	actualHub, ok := hub.(*SocketHub)
	if !ok {
		conn.Write([]byte("Bad Gate\n"))
		conn.Close()
		return
	}

	if !SecureWrite([]byte("Register successfully!\n"), conn, srv) {
		return
	}
	t.Service.Info <- conn.RemoteAddr().String() + " client register hub " + connName

	var newClient *SocketClient

	select {
	case binClient := <-t.ClientRecycler:
		newClient = binClient
	default:
		newClient = NewSocketClient()
		newClient.Listener = t
		newClient.Service = t.Service
	}
	newClient.Hub = actualHub
	socketClientCtx, socketClientCancel := context.WithCancel(actualHub.Context)
	newClient.Context = socketClientCtx
	newClient.Cancel = socketClientCancel

	go newClient.HandConn(conn)
	go newClient.Broadcast()
}
