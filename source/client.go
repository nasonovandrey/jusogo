package source

import (
	"bufio"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Client struct {
	name       string
	address    *net.UDPAddr
	connection *net.UDPConn
}

func CreateClient(name, remoteAddrString, localAddrString string) (*Client, error) {
	remoteAddr, err := net.ResolveUDPAddr("udp", remoteAddrString)
	if err != nil {
		return nil, err
	}
	localAddr, err := net.ResolveUDPAddr("udp", localAddrString)
	if err != nil {
		return nil, err
	}

	log.Printf("Creating connection from %s to %s.", localAddr.String(), remoteAddr.String())
	connection, err := net.DialUDP("udp", localAddr, remoteAddr)
	if err != nil {
		return nil, err
	}

	return &Client{
		name:       name,
		address:    localAddr,
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
		bytes_read, _, err := client.connection.ReadFromUDP(buffer)
		if err != nil {
			log.Println("Error reading from UDP:", err)
		} else {
			udpAddress := string(buffer[:bytes_read])
			peerAddress, err := net.ResolveUDPAddr("udp", udpAddress)
			if err != nil {
				log.Println("Error resolving UDP:", err, udpAddress)
			}
			client.connection.Close()
			EstablishP2PChat(client.address, peerAddress)
		}
	}
}

func EstablishP2PChat(localAddr, remoteAddr *net.UDPAddr) {
	log.Printf("Creating connection from %s to %s.", localAddr.String(), remoteAddr.String())
	conn, err := net.DialUDP("udp", localAddr, remoteAddr)
	if err != nil {
		log.Fatal(err)
	}

	go func(conn *net.UDPConn) {
		for {
			reader := bufio.NewReader(os.Stdin)
			text, _ := reader.ReadString('\n')
			conn.Write([]byte(text))
			time.Sleep(TIMEOUT)
		}
	}(conn)

	go func(conn *net.UDPConn) {
		buffer := make([]byte, 1024)
		for {
			bytes_read, _, _ := conn.ReadFromUDP(buffer)
			log.Printf(string(buffer[:bytes_read]))
		}
	}(conn)

	select {}

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
