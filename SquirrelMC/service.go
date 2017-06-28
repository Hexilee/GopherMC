package main

import "github.com/takama/daemon"

type Service struct {
	daemon.Daemon
	TCPHubListener *TCPHubListener
	TCPClientListener *TCPClientListener
	Signal chan string
	Config *ConfigType
}

func NewService(Config *ConfigType, srv daemon.Daemon) *Service {
	return &Service{
		Daemon: srv,
		Signal: make(chan string, 10000),
		Config: Config,
	}
}