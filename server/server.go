package server

import (
	"log"
	"net"
	"strings"
	"sync"
)

type Node struct {
	addr *net.UDPAddr
}

type Server struct {
	connection *net.UDPConn
	nodes      map[string]Node
	mutex      sync.Mutex
}

func CreateServer(addrString string) (*Server, error) {
	addr, err := net.ResolveUDPAddr("udp", addrString)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}
	log.Printf("Server listening on %s", addrString)
	return &Server{
		connection: conn,
		nodes:      make(map[string]Node),
	}, nil
}

func DeleteServer(server *Server) {
	log.Println("Shutting down the server.")
	server.connection.Close()
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
		clientName := strings.TrimSpace(message)

		if clientName == "" {
			log.Printf("Received empty message from %v. Ignoring.", remoteAddr)
			continue
		}

		server.mutex.Lock()
		server.nodes[clientName] = Node{addr: remoteAddr}
		server.mutex.Unlock()

		log.Printf("Registered client: %s from %v", clientName, remoteAddr)
	}
}
