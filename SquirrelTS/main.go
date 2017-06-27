package main

import (
	"flag"
	"net"
	"fmt"
)

var (
	service string
)

func main() {
	flag.StringVar(&service, "c", "0.0.0.0:8080", "the host and port you dial")
	flag.Parse()

	tcpAddr, _ := net.ResolveTCPAddr("tcp4", service)
	conn, _ := net.DialTCP("tcp", nil, tcpAddr)

	_, _ = conn.Write([]byte("xixi"))
	_, _ = conn.Write([]byte("haha"))

	var data []byte
	_, _ = conn.Read(data)
	fmt.Println(data)

}
