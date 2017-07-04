package main

import (
	"flag"
	"github.com/jinzhu/configor"
	"os"
	"fmt"
)

var (
	conf string
)

func main() {
	flag.StringVar(&conf, "f", "./config.conf", "the path config file, the default is ./config.conf")
	flag.Parse()

	configor.Load(&Config, conf)

	service := NewService(&Config)
	//go service.Logger(Config.LogFile)
	status, err := service.Start()
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
	fmt.Println(status)
}
