package main

import (
	"flag"
	"github.com/jinzhu/configor"
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

	service := NewService(&Config)
	//go service.Logger(Config.LogFile)
	status, err := service.Start()
	if err != nil {
		errlog.Println("Error: ", err)
		os.Exit(1)
	}
	fmt.Println(status)
}
