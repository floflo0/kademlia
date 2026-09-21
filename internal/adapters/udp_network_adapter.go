package adapters

import (
	"fmt"
	"kademlia/internal/core/entities"
	"kademlia/internal/core/ports"
	"net"
	"time"
)

type UdpNetworkAdapter struct{}

type udpListenConnection struct {
	connection *net.UDPConn
}

type udpDialConnection struct {
	connection *net.UDPConn
}

const maxMessageSize = 4096

func NewUDPNetworkAdapter() *UdpNetworkAdapter {
	return &UdpNetworkAdapter{}
}

func (*UdpNetworkAdapter) Listen(address entities.Address) (ports.ListenConnection, error) {
	addr, err := addressToUdpAdrr(address)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}
	connection := &udpListenConnection{
		connection: conn,
	}
	return connection, nil
}

func (*UdpNetworkAdapter) Dial(address entities.Address) (ports.DialConnection, error) {
	addr, err := addressToUdpAdrr(address)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, err
	}

	connection := &udpDialConnection{
		connection: conn,
	}
	return connection, nil
}

func (c *udpListenConnection) GetIP() (string, error) {
	addr := c.connection.LocalAddr().(*net.UDPAddr)
	if addr.IP.IsUnspecified() {
		conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
			IP:   net.ParseIP("1.1.1.1"),
			Port: 80,
		})
		if err != nil {
			return "", err
		}
		defer conn.Close()
		addr := conn.LocalAddr().(*net.UDPAddr)
		return addr.IP.String(), nil
	}
	return addr.IP.String(), nil
}

func (c *udpListenConnection) SendTo(address entities.Address, payload []byte) error {
	addr, err := addressToUdpAdrr(address)
	if err != nil {
		return err
	}
	_, err = c.connection.WriteToUDP(payload, addr)
	return err
}

func (c *udpListenConnection) Receive() ([]byte, *entities.Address, error) {
	buffer := make([]byte, maxMessageSize)
	n, addr, err := c.connection.ReadFromUDP(buffer)
	if err != nil {
		return nil, nil, err
	}
	address := udpAddrtoAddress(addr)
	payload := buffer[:n]
	return payload, &address, nil
}

func (c *udpListenConnection) Close() error {
	return c.connection.Close()
}

func (c *udpDialConnection) Send(payload []byte) error {
	_, err := c.connection.Write(payload)
	return err
}

func (c *udpDialConnection) Receive(timeoutMiliseconds uint32) ([]byte, error) {
	buffer := make([]byte, maxMessageSize)
	deadline := time.Now().Add(time.Duration(timeoutMiliseconds) * time.Millisecond)
	err := c.connection.SetReadDeadline(deadline)
	if err != nil {
		return nil, err
	}
	n, err := c.connection.Read(buffer)
	if err != nil {
		return nil, err
	}
	return buffer[:n], nil
}

func (c *udpDialConnection) Close() error {
	return c.connection.Close()
}

func addressToUdpAdrr(address entities.Address) (*net.UDPAddr, error) {
	ip := net.ParseIP(address.IP)
	if ip == nil {
		return nil, fmt.Errorf("failed to parse IP address: %q", address.IP)
	}
	addr := &net.UDPAddr{
		IP:   ip,
		Port: address.Port,
	}
	return addr, nil
}

func udpAddrtoAddress(addr *net.UDPAddr) entities.Address {
	return entities.Address{
		IP:   addr.IP.String(),
		Port: addr.Port,
	}
}
