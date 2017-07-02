package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"
	"bytes"
	"encoding/binary"
)

var host = flag.String("h", "127.0.0.1", "host")
var port = flag.String("p", "8080", "port")

func main() {
	flag.Parse()
	tcpAddr, err := net.ResolveTCPAddr("tcp4", *host + ":" + *port)
	if err != nil {
		fmt.Println("Error connecting:", err)
		os.Exit(1)
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Println("Error connecting:", err)
		os.Exit(1)
	}
	//defer conn.Close()
	fmt.Println("Connecting to " + *host + ":" + *port)
	done := make(chan []byte, 100000)
	register(conn, done)
	time.Sleep(time.Second)
	go handleRead(conn, done)
	handWrite(conn)

}

func handleRead(conn *net.TCPConn, done chan []byte) {
	for i := 0; i < 10; i++ {
		var data = make([]byte, 1024, 1024)
		_, _ = conn.Read(data)
		fmt.Println(string(data))
	}
}


func handWrite(conn *net.TCPConn) {
	start := time.Now()
	for i := 0; i < 3; i++ {
		data := []byte("xixi")
		head := make([]byte, 4)
		binary.LittleEndian.PutUint32(head, uint32(len(data)))
		buf := bytes.NewBuffer(head)
		buf.Write(data)
		conn.Write(buf.Bytes())
		time.Sleep(time.Second)
	}
	fmt.Println(time.Since(start).Seconds())
}


func register(conn *net.TCPConn, done chan []byte) {

	conn.Write([]byte("xixi"))
	//start := time.Now()
	//for i := 0; i < 100000; i++ {
	//	conn.Write([]byte("xixi"))
	//}
	//fmt.Println(time.Since(start).Seconds())
}
