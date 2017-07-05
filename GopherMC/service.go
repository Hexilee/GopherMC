package main

import (
	"os"
	"os/signal"
	"syscall"
	"context"
)



type Service struct {
	SocketHubListener    *SocketHubListener
	SocketClientListener *SocketClientListener
	Signal               chan string
	Config               *ConfigType
	Context              context.Context
	Cancel               context.CancelFunc
}

func NewService(Config *ConfigType) *Service {
	ctx, cancelFunc := context.WithCancel(context.Background())
	return &Service{
		Signal:  make(chan string, 10),
		Config:  Config,
		Context: ctx,
		Cancel:  cancelFunc,
	}
}

func (s *Service) Start() (string, error) {

	defer s.Cancel()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	socketHubCtx, socketHubCancel := context.WithCancel(s.Context)
	s.SocketHubListener = NewSocketHubListener(Config.Hub.Tcp.Service)
	s.SocketHubListener.Service = s
	s.SocketHubListener.Context = socketHubCtx
	s.SocketHubListener.Cancel = socketHubCancel

	socketClientCtx, socketClientCancel := context.WithCancel(s.Context)
	s.SocketClientListener = NewSocketClientListener(Config.Client.Tcp.Service)
	s.SocketClientListener.Service = s
	s.SocketClientListener.Context = socketClientCtx
	s.SocketClientListener.Cancel = socketClientCancel

	go s.SocketHubListener.Start()
	go s.SocketClientListener.Start(s.SocketHubListener.HubTable, s)

	for {
		select {
		case sysSignal := <-interrupt:
			Logger.Info("Got signal:" + sysSignal.String())
			switch sysSignal {
			case syscall.SIGINT:
				return "GopherMC was killed", nil
			case os.Interrupt:
				return "GopherMC was interrupted by system signal", nil
			case syscall.SIGTSTP:
				return "GopherMC cannot be put background", nil
			default:
				Logger.Info("Receive unknow signal: %v", sysSignal)
			}
			
		}
	}
}
