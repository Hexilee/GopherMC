package main

import (
	"io"
	"errors"
	"reflect"
)


func CheckErr(err error, srv *Service) bool {
	if err != nil {
		srv.Error <- &err
		return false
	}
	return true
}

func CheckPanic(_panic interface{}, srv *Service, errInfo string) bool {
	if _panic != nil {
		panicErr, ok := _panic.(error)
		if ok && reflect.TypeOf(panicErr).String() == "runtime.plainError" {
			srv.Error <- &panicErr
			err := errors.New(errInfo)
			srv.Error <- &err
			return false
		}
		panic(_panic)
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