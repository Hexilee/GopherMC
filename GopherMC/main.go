package main

import (
	"flag"
	"github.com/jinzhu/configor"
	"github.com/takama/daemon"
	"os"
	"log"
	"fmt"
)

var (
	conf           string
	stdlog, errlog *log.Logger
)

func init() {
	stdlog = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	errlog = log.New(os.Stderr, "", log.Ldate|log.Ltime)
}

func main() {

	flag.StringVar(&conf, "f", "./config.conf", "the path config file, the default is ./config.conf")
	flag.Parse()

	configor.Load(&Config, conf)

	srv, err := daemon.New(Config.APPName, Config.Description)

	if err != nil {
		errlog.Println("Error: ", err)
		os.Exit(1)
	}

	service := NewService(&Config, srv)
	//go service.Logger(Config.LogFile)
	status, err := service.Manage()
	if err != nil {
		errlog.Println("Error: ", err)
		os.Exit(1)
	}
	fmt.Println(status)
}
