package main

import (
	"os"
	"github.com/op/go-logging"
)

var (
	Logger = logging.MustGetLogger("example")
	format = logging.MustStringFormatter(
		`%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x} %{message}`,
	)
	logFile *os.File
	err error
)

