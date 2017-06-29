package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

var host = flag.String("host", "127.0.0.1", "host")
var port = flag.String("port", "8080", "port")

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
	handleWrite(conn, done)
	go handleRead(conn, done)
}

func handleRead(conn *net.TCPConn, done chan []byte) {
	for {
		var data = make([]byte, 1024, 1024)
		_, _ = conn.Read(data)
		fmt.Println(data)
	}
}

func handleWrite(conn *net.TCPConn, done chan []byte) {

	conn.Write([]byte("xixi"))
	start := time.Now()
	for i := 0; i < 100000; i++ {
		conn.Write([]byte("xixi"))
	}
	fmt.Println(time.Since(start).Seconds())
}
