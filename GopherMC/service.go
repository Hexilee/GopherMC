package main

import (
	"github.com/takama/daemon"
	"os"
	"github.com/op/go-logging"
	"os/signal"
	"syscall"
)

var (
	logger = logging.MustGetLogger("example")
	// Example format string. Everything except the message has a custom color
	// which is dependent on the log level. Many fields have a custom output
	// formatting too, eg. the time returns the hour down to the milli second.
	format = logging.MustStringFormatter(
		`%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x} %{message}`,
	)
)

type Service struct {
	daemon.Daemon
	SocketHubListener    *SocketHubListener
	SocketClientListener *SocketClientListener
	Signal               chan string
	Info                 chan string
	Error                chan *error
	Config               *ConfigType
}

func (s *Service) Logger(logfile string) {
	logFile, err := os.OpenFile(logfile, os.O_WRONLY, 0666)
	if err != nil {
		logger.Critical("Cannot open log file\n Error: ", err)
		os.Exit(1)
	}

	defer logFile.Close()
	backend1 := logging.NewLogBackend(logFile, "", 0)
	backend1Formatter := logging.NewBackendFormatter(backend1, format)
	logging.SetBackend(backend1Formatter)
	logger.Info("Start")
	defer logger.Info("Stop")
	for {
		select {
		case info := <-s.Info:
			logger.Info(info)
		case err := <-s.Error:
			logger.Error((*err).Error())
		case Signal := <-s.Signal:
			if Signal == "kill" {
				logger.Critical("Signal killed!")
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

func (s *Service) Manage() (string, error) {

	usage := "Usage: GopherMC restart | start | stop | status"

	// if received any kind of command, do it
	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "install":
			return s.Install()
		case "remove":
			return s.Remove()
		case "restart":
			return s.Restart()
		case "start":
			return s.Start()
		case "stop":
			return s.Stop()
		case "status":
			return s.Status()
		default:
			return usage, nil
		}
	}

	// Do something, call your goroutines, etc

	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	// Set up listener for defined host and port

	// set up channel on which to send accepted connections
	s.SocketHubListener = NewSocketHubListener(Config.Hub.Tcp.Service, s)
	s.SocketHubListener.Service = s

	s.SocketClientListener = NewSocketClientListener(Config.Client.Tcp.Service, s)
	s.SocketClientListener.Service = s
	////go ListenHub(Config.Hub.Tcp.MaxBytes, Config.Hub.Tcp.Service, &hubMap)
	go s.SocketHubListener.Start(Config.Hub.Tcp.MaxBytes, s)
	go s.SocketClientListener.Start(Config.Client.Tcp.MaxBytes, s.SocketHubListener.HubTable, s)
	go s.Logger(Config.LogFile)
	// loop work cycle with accept connections or interrupt
	// by system signal
	for {
		select {
		case killSignal := <-interrupt:
			s.Info <- "Got signal:" + killSignal.String()
			if killSignal == os.Interrupt {
				return "Daemon was interruped by system signal", nil
			}
			return "Daemon was killed", nil
		}
	}
}

//func (s *Service) Start() (string, error) {
//	if len(os.Args) > 3 {
//		Type := os.Args[2]
//		protocal := os.Args[3]
//		if Type == "client" {
//			switch protocal {
//			case "socket":
//				if s.SocketHubListener == nil {
//					return "SocketHubListener is closed", nil
//				}
//				s.SocketClientListener = NewSocketClientListener(Config.Client.Tcp.Service)
//				s.SocketClientListener.Service = s
//				s.SocketClientListener.Start(Config.Hub.Tcp.MaxBytes, s.SocketHubListener.HubTable)
//				return "TCP socket client started at " + Config.Client.Tcp.Service, nil
//			}
//		} else if Type == "hub" {
//			switch protocal {
//			case "socket":
//				s.SocketHubListener = NewSocketHubListener(Config.Hub.Tcp.Service)
//				s.SocketHubListener.Service = s
//				s.SocketHubListener.Start(Config.Hub.Tcp.MaxBytes)
//				return "TCP socket hub started at " + Config.Hub.Tcp.Service, nil
//			}
//		}
//	}
//	return "Use GopherMC start [service type] socket | ws", nil
//}

func (s *Service) Restart() (string, error) {
	return "Use GopherMC restart [service type] socket | ws", nil
}
