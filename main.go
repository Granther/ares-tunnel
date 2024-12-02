package main

import (
	"fmt"
	"glorpn/client"
	"glorpn/server"
	"log"
	"os"
)

func main() {
	args := os.Args

	if len(args) <= 1 {
		log.Fatalln("Incorrect usage")
	} else if args[1] == "client" {
		fmt.Println("Running Client")
		var wanIface string; var peerIP string
		if len(args) < 3 {
			wanIface = "eth0"
		} else {
			wanIface = args[2]
		}

		if len(args) < 4 {
			peerIP = ""
		} else {
			peerIP = args[3]
		}
		
		err := runClient(wanIface, peerIP)
		if err != nil {
			log.Fatalln("Error in runclient: ", err)
		}
	} else if args[1] == "server" {
		fmt.Println("Running Server")
		err := runServer()
		if err != nil {
			log.Fatalln("Error in runserver: ", err)
		}
	} else {
		log.Fatalln("Unknown argument")
	}
}

func runServer() error {
	server := server.NewServer()
	return server.Start()
}

func runClient(wanIface, peerIP string) error {
	client := client.NewClient()
	return client.Start(wanIface, peerIP)
}
