package source

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Node struct {
	addr     *net.UDPAddr
	lastSeen time.Time
}

func (node Node) String() string {
	return fmt.Sprintf("address=%s lastSeen=%s", node.addr, node.lastSeen.Format("2006-01-02T15:04:00"))
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

func UpdateNodes(server *Server) {
	buffer := make([]byte, 1024)
	for {
		bytesRead, remoteAddr, err := server.connection.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Error reading from UDP: %v", err)
			continue
		}

		message := string(buffer[:bytesRead])
		nodeName := strings.TrimSpace(message)

		if nodeName == "" {
			log.Printf("Received empty message from %v. Ignoring.", remoteAddr)
			continue
		}

		server.mutex.Lock()
		server.nodes[nodeName] = Node{addr: remoteAddr, lastSeen: time.Now()}
		log.Printf("Registered client: %s from %v", nodeName, remoteAddr)
		server.mutex.Unlock()
	}
}

func EvictNodes(server *Server) {
	for {
		server.mutex.Lock()
		for nodeName, node := range server.nodes {
			if time.Now().Sub(node.lastSeen) > TIMEOUT {
				delete(server.nodes, nodeName)
				log.Printf("Client %s last seen %v. Removing.", nodeName, node.lastSeen)
			}
		}
		server.mutex.Unlock()
		time.Sleep(TIMEOUT)
	}
}

func RunHttp(server *Server) {
	http.HandleFunc("/query", func(writer http.ResponseWriter, request *http.Request) {
		var output string
		for nodeName, node := range server.nodes {
			output += fmt.Sprintf("nodeName=%s %s\n", nodeName, node)
		}
		writer.Write([]byte(output))
	})
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func RunServer(server *Server) {
	log.Println("Server is running and ready to receive data.")

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// These things spam on shutdown. Perhaps they
	// should read from a channel too and shutdown gracefully
	go UpdateNodes(server)
	go EvictNodes(server)

	// Separate web service for querying
	go RunHttp(server)

	<-done

	log.Printf("Server is shutting down.")

}
