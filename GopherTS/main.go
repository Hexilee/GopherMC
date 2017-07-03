package main

import (
	"os"
)



func main() {
	if len(os.Args) > 2 {
		command := os.Args[1]
		switch command {
		case "hub":
			testHub := NewTestHub()
			testHub.StartHub()
		case "client":
			testClient := NewTestClient()
			testClient.StartClient()
		}
	}
}

//func handleRead(conn *net.TCPConn, done chan []byte) {
//	for i := 0; i < 10; i++ {
//		var data = make([]byte, 1024, 1024)
//		_, _ = conn.Read(data)
//		fmt.Println(string(data))
//	}
//}
//
//func handWrite(conn *net.TCPConn) {
//	start := time.Now()
//	for i := 0; i < 3; i++ {
//		data := []byte("xixi")
//		head := make([]byte, 4)
//		binary.LittleEndian.PutUint32(head, uint32(len(data)))
//		buf := bytes.NewBuffer(head)
//		buf.Write(data)
//		conn.Write(buf.Bytes())
//		time.Sleep(time.Second)
//	}
//	fmt.Println(time.Since(start).Seconds())
//}
//
//func register(conn *net.TCPConn, done chan []byte) {
//
//	conn.Write([]byte("xixi"))
//	//start := time.Now()
//	//for i := 0; i < 100000; i++ {
//	//	conn.Write([]byte("xixi"))
//	//}
//	//fmt.Println(time.Since(start).Seconds())
//}
