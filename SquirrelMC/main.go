package main

import (
	S "golang.org/x/sync/syncmap"
	"flag"
)


var (
	clientService string
	hubService string
	byteMax int
	hubMap = S.Map {}
)


func main() {

	flag.StringVar(&clientService, "c", "0.0.0.0:8080", "the host and port your client listener bind")
	flag.StringVar(&hubService, "h", "0.0.0.0:8888", "the host and port your hub listener bind")
	flag.IntVar(&byteMax, "b", 512, "the maximum of the length of message")
	flag.Parse()

	go ListenHub(byteMax, hubService, &hubMap)
	ListenClient(byteMax, clientService, &hubMap)
}