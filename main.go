package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"jusogo/source" 
)

func main() {
	host := flag.String("host", "0.0.0.0", "The IP address the server binds to")
	port := flag.Int("port", 7000, "The port the server listens on")
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Println("Usage: server --host <host> --port <port> <command>")
		fmt.Println("Commands:")
		fmt.Println("  start   Start the server")
		os.Exit(1)
	}

	command := flag.Args()[0]
	switch command {
	case "start":
		srv, err := server.CreateServer(*host, *port)
		if err != nil {
			log.Fatal(err)
		}
		defer server.DeleteServer(srv)
		server.RunServer(srv)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Available commands: start")
		os.Exit(1)
	}
}

