package main

import (
	"io"
	"encoding/binary"
	"bufio"
	"bytes"
	"errors"
)

const (
	headerLen int = 4
)

func CheckErr(err error, srv *Service) bool {
	if err != nil {
		srv.Error <- &err
		return false
	}
	return true
}

func CheckPanic(panic interface{}, srv *Service, errInfo string) bool {
	if panic != nil {
		err := errors.New(errInfo)
		srv.Error <- &err
		return false
	}
	return true
}

func DealConnErr(err error, closer io.Closer, srv *Service) bool {
	if err != nil {
		closer.Close()
		srv.Error <- &err
		return false
	}
	return true
}

func SecureWrite(msg []byte, writeCloser io.WriteCloser, srv *Service) bool {
	_, err := writeCloser.Write(msg)
	return DealConnErr(err, writeCloser, srv)
}

func SecureRead(msg []byte, ReadCloser io.ReadCloser, srv *Service) bool {
	_, err := ReadCloser.Read(msg)
	return DealConnErr(err, ReadCloser, srv)
}

//func ListenHub(bytes int, service string, hubtable *S.Map) {
//	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
//	CheckErr(err)
//
//	listener, err := net.ListenTCP("tcp", tcpAddr)
//	CheckErr(err)
//
//	log.Printf("Hub listener listening at %s", service)
//
//	for {
//		conn, err := listener.AcceptTCP()
//		CheckErr(err)
//		conn.Write([]byte("Fine, connected\n")) // don't care about return value conn.Close()
//		go func() {
//			var registerInfo [32]byte
//			conn.Read(registerInfo[:])
//			connName := string(registerInfo[:])
//
//			newHub := NewTCPHub()
//			_, ok := hubtable.LoadOrStore(connName, newHub)
//			if ok {
//				conn.Write([]byte("The hub already exist!\n"))
//				conn.Close()
//				return
//			}
//			newHub.Conn = conn
//			_, err = conn.Write([]byte("Register successfully!\n"))
//			DealConnErr(err, conn)
//			go newHub.HandConn(conn, bytes)
//			go newHub.RegisterClient()
//			go newHub.SendMessage()
//			go newHub.ClientWriter()
//		}()
//	}
//}
//
//func ListenClient(bytes int, service string, hubtable *S.Map) {
//	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
//	CheckErr(err)
//
//	listener, err := net.ListenTCP("tcp", tcpAddr)
//	CheckErr(err)
//
//	log.Printf("Client listener listening at %s", service)
//
//	for {
//		conn, err := listener.AcceptTCP()
//		CheckErr(err)
//		conn.Write([]byte("Fine, connected\n")) // don't care about return value conn.Close()
//
//		go func() {
//			var registerInfo [32]byte
//			conn.Read(registerInfo[:])
//			connName := string(registerInfo[:])
//
//			hub, ok := hubtable.Load(connName)
//			if !ok {
//				conn.Write([]byte("No hub named!" + connName))
//				conn.Close()
//				return
//			}
//			actualHub, ok := hub.(*TCPHub)
//			if !ok {
//				conn.Write([]byte("Bad Gate\n"))
//				conn.Close()
//				return
//			}
//
//			newClient := NewTCPClient()
//			conn.Write([]byte("Register successfully!\n"))
//			go newClient.HandConn(conn, bytes, actualHub)
//			go newClient.Broadcast()
//		}()
//	}
//}



func SocketRead(conn io.ReadWriteCloser, ch chan []byte, serv *Service) {
	scanner := bufio.NewScanner(conn)
	split := func(data []byte, atEOF bool) (adv int, token []byte, err error) {
		length := len(data)
		if length < headerLen {
			return 0, nil, nil
		}
		if length > 1048576 { //1024*1024=1048576
			conn.Close()
			serv.Info <- "invalid query!"
			return 0, nil, errors.New("too large data!")
		}
		var lhead uint32
		buf := bytes.NewReader(data)
		binary.Read(buf, binary.LittleEndian, &lhead)
		tail := length - headerLen
		if lhead > 1048576 {
			conn.Close()
			serv.Info <- "invalid query2!"
			return 0, nil, errors.New("too large data!")
		}
		if uint32(tail) < lhead {
			return 0, nil, nil
		}
		adv = headerLen + int(lhead)
		token = data[:adv]
		return
	}
	scanner.Split(split)
	for scanner.Scan() {
		var msg []byte
		copy(msg, scanner.Bytes())
		ch <- msg
	}
	if scanner.Err() != nil {
		err := scanner.Err()
		serv.Error <- &err
	}
}
