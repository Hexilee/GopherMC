package main

import (
	"flag"
	"github.com/jinzhu/configor"
	"github.com/takama/daemon"
	"os"
)


var (
	conf string
)



func main() {

	flag.StringVar(&conf, "f", "./config.conf", "the path config file, the default is ./config.conf")
	flag.Parse()

	configor.Load(&Config, conf)

	srv, err := daemon.New(Config.APPName, Config.Description)

	if err != nil {
		log.Critical("Cannot create daemon\n Error: ", err)
		os.Exit(1)
	}

	service := NewService(&Config, srv)

	service.TCPHubListener = NewTCPHubListener(Config.Hub.Tcp.Service)
	service.TCPHubListener.Service = service

	service.TCPClientListener = NewTCPClientListener(Config.Client.Tcp.Service)
	service.TCPClientListener.Service = service
	//go ListenHub(Config.Hub.Tcp.MaxBytes, Config.Hub.Tcp.Service, &hubMap)
	go service.TCPHubListener.Start(Config.Hub.Tcp.MaxBytes)
	go service.TCPClientListener.Start(Config.Client.Tcp.MaxBytes, service.TCPHubListener.HubTable)
	service.Logger(Config.LogFile)
}