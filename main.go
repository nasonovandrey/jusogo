package main

import (
	"flag"
	"fmt"
	"jusogo/source"
	"log"
	"os"
)

func main() {
	mode := flag.String("mode", "", "Specify 'server' or 'client'")
	host := flag.String("host", "localhost", "Specify the hostname")
	port := flag.Int("port", 7000, "Specify the port number")
	name := flag.String("name", "", "Specify the client name (required for client mode)")
	localport := flag.Int("localport", 0, "Specify the local port for the client (required for client mode)")

	flag.Parse()

	if *mode != "server" && *mode != "client" {
		fmt.Println("Usage: -mode server|client -host <hostname> -port <port> [-name <client name>]")
		os.Exit(1)
	}

	serverAddress := fmt.Sprintf("%s:%d", *host, *port)

	if *mode == "server" {
		srv, err := source.CreateServer(serverAddress)
		if err != nil {
			log.Fatal(err)
		}
		defer source.DeleteServer(srv)
		source.RunServer(srv)
	} else if *mode == "client" {
		if *name == "" {
			log.Fatal("Error: -name flag must be specified for client mode")
		}
		if *localport == 0 {
			log.Fatal("Error: -localport flag must be specified for client mode")
		}

		cln, err := source.CreateClient(*name, serverAddress, fmt.Sprintf("localhost:%d", *localport))
		if err != nil {
			log.Fatal(err)
		}
		defer source.DeleteClient(cln)
		source.RunClient(cln)
	}
}
