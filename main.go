package main

import (
	"flag"
	"fmt"
	"jusogo/client"
	"jusogo/server"
	"log"
	"os"
)

func main() {
	mode := flag.String("mode", "", "Specify 'server' or 'client'")
	host := flag.String("host", "localhost", "Specify the hostname")
	port := flag.Int("port", 7000, "Specify the port number")
	name := flag.String("name", "", "Specify the client name (required for client mode)")

	flag.Parse()

	if *mode != "server" && *mode != "client" {
		fmt.Println("Usage: -mode server|client -host <hostname> -port <port> [-name <client name>]")
		os.Exit(1)
	}

	serverAddress := fmt.Sprintf("%s:%d", *host, *port)

	if *mode == "server" {
		srv, err := server.CreateServer(serverAddress)
		if err != nil {
			log.Fatal(err)
		}
		defer server.DeleteServer(srv)
		server.RunServer(srv)
	} else if *mode == "client" {
		if *name == "" {
			log.Fatal("Error: -name flag must be specified for client mode")
		}

		cln, err := client.CreateClient(*name, serverAddress)
		if err != nil {
			log.Fatal(err)
		}
		defer client.DeleteClient(cln)
		client.RunClient(cln)
	}
}
