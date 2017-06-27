package main

import (
	"fmt"
	"net"
	S "golang.org/x/sync/syncmap"
	"log"
)

func CheckErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func ListenHub(bytes int, service string, hubtable *S.Map) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	CheckErr(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	CheckErr(err)

	log.Printf("Hub listener listening at %s", service)

	for {
		conn, err := listener.AcceptTCP()
		CheckErr(err)
		conn.Write([]byte("Fine, connected\n")) // don't care about return value conn.Close()
		go func() {
			var registerInfo [32]byte
			conn.Read(registerInfo[:])
			connName := string(registerInfo[:])

			newHub := NewTCPHub()
			_, ok := hubtable.LoadOrStore(connName, newHub)
			if ok {
				conn.Write([]byte("The hub already exist!\n"))
				conn.Close()
				return
			}
			newHub.Conn = conn
			conn.Write([]byte("Register successfully!\n"))
			go newHub.HandConn(conn, bytes)
			go newHub.RegisterClient()
			go newHub.SendMessage()
			go newHub.ClientWriter()
		}()
	}
}

func ListenClient(bytes int, service string, hubtable *S.Map) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	CheckErr(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	CheckErr(err)

	log.Printf("Client listener listening at %s", service)

	for {
		conn, err := listener.AcceptTCP()
		CheckErr(err)
		conn.Write([]byte("Fine, connected\n")) // don't care about return value conn.Close()

		go func() {
			var registerInfo [32]byte
			conn.Read(registerInfo[:])
			connName := string(registerInfo[:])

			hub, ok := hubtable.Load(connName)
			if !ok {
				conn.Write([]byte("No hub named!" + connName))
				conn.Close()
				return
			}
			actualHub, ok := hub.(*TCPHub)
			if !ok {
				conn.Write([]byte("Bad Gate\n"))
				conn.Close()
				return
			}

			newClient := NewTCPClient()
			conn.Write([]byte("Register successfully!\n"))
			go newClient.HandConn(conn, bytes, actualHub)
			go newClient.Broadcast()
		}()
	}
}
