package main

import (
	"flag"
	"github.com/jinzhu/configor"
	"github.com/op/go-logging"
	"os"
	"fmt"
)

var (
	conf string
)


func init() {

	flag.StringVar(&conf, "f", "./config.conf", "the path config file, the default is ./config.conf")
	flag.Parse()

	configor.Load(&Config, conf)

	logFile, err = os.OpenFile(Config.LogFile, os.O_WRONLY, 0666)
	if err != nil {
		Logger.Critical("Cannot open log file\n Error: ", err)
		os.Exit(1)
	}

	backend1 := logging.NewLogBackend(logFile, "", 0)
	backend1Formatter := logging.NewBackendFormatter(backend1, format)
	logging.SetBackend(backend1Formatter)
}

func main() {

	defer logFile.Close()
	service := NewService(&Config)
	//go service.Logger(Config.LogFile)
	status, err := service.Start()
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
	fmt.Println(status)
}
