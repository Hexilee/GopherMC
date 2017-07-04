package main

import (
	"os"
	"github.com/op/go-logging"
	"os/signal"
	"syscall"
	"context"
)

var (
	logger = logging.MustGetLogger("example")
	format = logging.MustStringFormatter(
		`%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x} %{message}`,
	)
)

type Service struct {
	SocketHubListener    *SocketHubListener
	SocketClientListener *SocketClientListener
	Signal               chan string
	Info                 chan string
	Error                chan *error
	Config               *ConfigType
	Context              context.Context
	Cancel               context.CancelFunc
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

func NewService(Config *ConfigType) *Service {
	ctx, cancelFunc := context.WithCancel(context.Background())
	return &Service{
		Signal:  make(chan string, 10),
		Info:    make(chan string, 100000),
		Error:   make(chan *error, 10000),
		Config:  Config,
		Context: ctx,
		Cancel:  cancelFunc,
	}
}

func (s *Service) Start() (string, error) {

	defer func() {
		s.Cancel()
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	socketHubCtx, socketHubCancel := context.WithCancel(s.Context)
	s.SocketHubListener = NewSocketHubListener(Config.Hub.Tcp.Service, s)
	s.SocketHubListener.Service = s
	s.SocketHubListener.Context = socketHubCtx
	s.SocketHubListener.Cancel = socketHubCancel

	socketClientCtx, socketClientCancel := context.WithCancel(s.Context)
	s.SocketClientListener = NewSocketClientListener(Config.Client.Tcp.Service, s)
	s.SocketClientListener.Service = s
	s.SocketClientListener.Context = socketClientCtx
	s.SocketClientListener.Cancel = socketClientCancel

	go s.SocketHubListener.Start(s)
	go s.SocketClientListener.Start(s.SocketHubListener.HubTable, s)
	go s.Logger(Config.LogFile)

	for {
		select {
		case killSignal := <-interrupt:
			s.Info <- "Got signal:" + killSignal.String()
			if killSignal == os.Interrupt {
				return "GopherMC was interrupted by system signal", nil
			}
			return "GopherMC was killed", nil
		}
	}
}
