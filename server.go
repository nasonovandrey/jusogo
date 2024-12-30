package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type Node struct {
	addr *net.UDPAddr
}

type Server struct {
	nodes      map[string]*Node
	connection *net.UDPConn
	mutex      sync.Mutex
}

func CreateServer(ip string, port int) (*Server, error) {
	udpAddr := &net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(ip),
	}
	connection, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}
	return &Server{
		nodes:      make(map[string]*Node),
		connection: connection,
	}, nil
}

func RunServer(server *Server) {
	buffer := make([]byte, 1024)
	log.Println("Server is running and ready to receive data.")

	for {
		bytesRead, remoteAddr, err := server.connection.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Error reading from UDP: %v", err)
			continue
		}

		message := string(buffer[:bytesRead])

		// No meaningful data yet
		clientName := strings.TrimSpace(message)

		if clientName == "" {
			log.Printf("Received empty message from %v. Ignoring.", remoteAddr)
			continue
		}

		server.mutex.Lock()
		server.nodes[clientName] = &Node{addr: remoteAddr}
		server.mutex.Unlock()

		log.Printf("[%s] Registered client: %s from %v", time.Now().Format(time.RFC3339), clientName, remoteAddr)
	}
}

func DeleteServer(server *Server) {
	server.connection.Close()
}

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
		server, err := CreateServer(*host, *port)
		if err != nil {
			log.Fatal(err)
		}
		defer DeleteServer(server)
		RunServer(server)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Available commands: start")
		os.Exit(1)
	}
}
