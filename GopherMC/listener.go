package main

import (
	"net"
	S "golang.org/x/sync/syncmap"
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

func NewSocketHubListener(service string) *SocketHubListener {

	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	if !CheckErr(err) {
		return nil
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if !CheckErr(err) {
		return nil
	}

	//logger.Info("Hub listener listening at %s", service)
	Logger.Info("Hub listener listening at %s", service)

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

func NewSocketClientListener(service string) *SocketClientListener {

	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	if !CheckErr(err) {
		return nil
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if !CheckErr(err) {
		return nil
	}
	//logger.Info("Client listener listening at %s", service)
	Logger.Info("Client listener listening at %s", service)

	return &SocketClientListener{
		Listener:       listener,
		Signal:         make(chan string, 5),
		ClientRecycler: make(chan *SocketClient, 10000),
		//Context:        context.Background(),
		//Cancel:         func() {},
	}
}

func (t *SocketHubListener) Start() {
	for {
		conn, err := t.Listener.AcceptTCP()
		if !CheckErr(err) {
			continue
		}
		go t.HandConn(conn)
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

func (t *SocketHubListener) HandConn(conn *net.TCPConn) {
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
	if !SecureWrite([]byte("Register successfully!\n"), conn) {
		return
	}
	Logger.Info("%s hub register as %s", conn.RemoteAddr().String(), connName)
	go newHub.Start(conn)
}

func (t *SocketClientListener) Start(HubTable *S.Map) {
	for {
		conn, err := t.Listener.AcceptTCP()
		if !CheckErr(err) {
			continue
		}
		go t.HandConn(conn, HubTable)
	}
}

func (t *SocketClientListener) HandConn(conn *net.TCPConn, HubTable *S.Map) {
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

	if !SecureWrite([]byte("Register successfully!\n"), conn) {
		return
	}
	Logger.Info("%s client register hub %s", conn.RemoteAddr().String(), connName)

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
