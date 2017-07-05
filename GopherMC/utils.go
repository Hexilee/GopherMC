package main

import (
	"io"
	"errors"
	"reflect"
)

func CheckErr(err error) bool {
	if err != nil {
		Logger.Error(err.Error())
		return false
	}
	return true
}

func CheckPanic(_panic interface{}, errInfo string) bool {
	if _panic != nil {
		panicErr, ok := _panic.(error)
		if ok && reflect.TypeOf(panicErr).String() == "runtime.plainError" {
			Logger.Error(panicErr.Error())
			err := errors.New(errInfo)
			Logger.Error(err.Error())
			return false
		}
		panic(_panic)
	}
	return true
}

func DealConnErr(err error, closer io.Closer) bool {
	if err != nil {
		closer.Close()
		Logger.Error(err.Error())
		return false
	}
	return true
}

func SecureWrite(msg []byte, writeCloser io.WriteCloser) bool {
	_, err := writeCloser.Write(msg)
	return DealConnErr(err, writeCloser)
}
