package source

import (
	"net"
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

func RunClient(client *Client) {
	for {
		client.connection.Write([]byte(client.name))
		time.Sleep(HEARTBEAT)
	}
}
