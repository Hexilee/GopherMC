package main

import (
	"net"
	"fmt"
	"os"
	"bufio"
	"encoding/binary"
	"bytes"
)


type TestClient struct {
}


func (t *TestClient) StartClient() {
	service := os.Args[2]
	conn, err := net.Dial("tcp", service)
	if err != nil {
		fmt.Println("Error connecting:", err)
		os.Exit(1)
	}
	defer conn.Close()
	fmt.Println("Connecting to " + service)
	go t.ReadConn(conn)
	t.WriteConn(conn)
}

func (t *TestClient) ReadConn(conn net.Conn) {
	buf := make([]byte, 10240)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Println(string(buf[:n]))
	}
}

func (t *TestClient) WriteConn(conn net.Conn) {
	for {
		bio := bufio.NewReader(os.Stdin)
		line, _, err := bio.ReadLine()
		if err != nil {
			fmt.Println(err)
			break
		}
		length := uint32(len(line))
		head := make([]byte, 4)
		binary.LittleEndian.PutUint32(head, length)
		buf := bytes.NewBuffer(head)
		buf.Write(line)
		_, err = conn.Write(buf.Bytes())
		if err != nil {
			fmt.Println(err)
			break
		}
	}
}

func NewTestClient() *TestClient {
	return &TestClient{}
}