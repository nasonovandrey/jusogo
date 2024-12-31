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
	address  *net.UDPAddr
	lastSeen time.Time
}

func (node Node) String() string {
	return fmt.Sprintf("address=%s lastSeen=%s", node.address, node.lastSeen.Format("2006-01-02T15:04:00"))
}

type Server struct {
	connection *net.UDPConn
	nodes      map[string]Node
	mutex      sync.Mutex
}

func CreateServer(addrString string) (*Server, error) {
	address, err := net.ResolveUDPAddr("udp", addrString)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", address)
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
		server.nodes[nodeName] = Node{address: remoteAddr, lastSeen: time.Now()}
		log.Printf("Registered node: %s from %v", nodeName, remoteAddr)
		server.mutex.Unlock()
	}
}

func EvictNodes(server *Server) {
	for {
		server.mutex.Lock()
		for nodeName, node := range server.nodes {
			if time.Now().Sub(node.lastSeen) > TIMEOUT {
				delete(server.nodes, nodeName)
				log.Printf("Node %s last seen %v. Removing.", nodeName, node.lastSeen)
			}
		}
		server.mutex.Unlock()
		time.Sleep(TIMEOUT)
	}
}

func InitiateP2P(server *Server, nodeNameX, nodeNameY string) (bool, error) {
	nodeX := server.nodes[nodeNameX]
	nodeY := server.nodes[nodeNameY]
	connectionX, err := net.DialUDP("udp", nil, nodeX.address)
	if err != nil {
		return false, err
	}
	connectionY, err := net.DialUDP("udp", nil, nodeY.address)
	if err != nil {
		return false, err
	}
	connectionX.Write([]byte(nodeY.address.String()))
	connectionY.Write([]byte(nodeX.address.String()))
	connectionX.Close()
	connectionY.Close()
	return true, nil
}

func RunHttp(server *Server) {
	http.HandleFunc("/query", func(writer http.ResponseWriter, request *http.Request) {
		var output string
		for nodeName, node := range server.nodes {
			output += fmt.Sprintf("nodeName=%s %s\n", nodeName, node)
		}
		writer.Write([]byte(output))
	})
	http.HandleFunc("/connect", func(writer http.ResponseWriter, request *http.Request) {
		request.ParseForm()
		nodeNameX := request.FormValue("X")
		nodeNameY := request.FormValue("Y")
		log.Printf("%s %s\n", nodeNameX, nodeNameY)
		InitiateP2P(server, nodeNameX, nodeNameY)
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
	// This is highly optional at this point, I'm not sure how
	// connection between nodes should be initiated at this point
	go RunHttp(server)

	<-done

	log.Printf("Server is shutting down.")

}
