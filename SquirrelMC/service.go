package main

import (
	"github.com/takama/daemon"
	"os"
	"github.com/op/go-logging"
)

var (
	log = logging.MustGetLogger("example")
	// Example format string. Everything except the message has a custom color
	// which is dependent on the log level. Many fields have a custom output
	// formatting too, eg. the time returns the hour down to the milli second.
	format = logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)
)

type Service struct {
	daemon.Daemon
	TCPHubListener    *TCPHubListener
	TCPClientListener *TCPClientListener
	Signal            chan string
	Info              chan string
	Error             chan *error
	Config            *ConfigType
}

func (s *Service) Logger(logfile string) {
	logFile, err := os.OpenFile(logfile, os.O_WRONLY, 0666)
	if err != nil {
		log.Critical("Cannot open log file\n Error: ", err)
		os.Exit(1)
	}

	defer logFile.Close()
	backend1 := logging.NewLogBackend(logFile, "", 0)
	backend1Formatter := logging.NewBackendFormatter(backend1, format)
	logging.SetBackend(backend1Formatter)
	log.Info("Start")
	defer log.Info("Stop")
	for {
		select {
		case info := <-s.Info:
			log.Info(info)
		case err := <-s.Error:
			log.Error((*err).Error())
		case signal := <-s.Signal:
			if signal == "kill" {
				log.Critical("Signal killed!")
			}
		}
	}
}

func NewService(Config *ConfigType, srv daemon.Daemon) *Service {
	return &Service{
		Daemon: srv,
		Signal: make(chan string, 10000),
		Info:   make(chan string, 100000),
		Error:  make(chan *error, 10000),
		Config: Config,
	}
}
