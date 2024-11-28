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
		err := runClient()
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

func runClient() error {
	client := client.NewClient()
	return client.Start()
}
