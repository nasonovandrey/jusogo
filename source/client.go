package source

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Client struct {
	name       string
	connection *net.UDPConn
}

func CreateClient(name, addrString string) (*Client, error) {
	addr, err := net.ResolveUDPAddr("udp", addrString)
	if err != nil {
		return nil, err
	}

	connection, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, err
	}

	return &Client{
		name:       name,
		connection: connection,
	}, nil

}

func DeleteClient(client *Client) {
	client.connection.Close()
}

func RegisterClient(client *Client) {
	for {
		client.connection.Write([]byte(client.name))
		time.Sleep(HEARTBEAT)
	}
}

func ReadFromServer(client *Client) {
	buffer := make([]byte, 1024)
	for {
		bytes_read, addr, err := client.connection.ReadFromUDP(buffer)
		if err != nil {
			log.Println("Error reading from UDP:", err)
		} else {
			log.Printf("Received message from %s: %s\n", addr.String(), string(buffer[:bytes_read]))
		}
	}
}

func RunClient(client *Client) {
	log.Println("Client is running.")

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go RegisterClient(client)
	go ReadFromServer(client)

	<-done
	log.Printf("Client is shutting down.")
}
