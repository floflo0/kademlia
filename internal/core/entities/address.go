package entities

import (
	"fmt"
	"net"
	"strconv"
)

type Address struct {
	IP   string
	Port int
}

func NewAddressFromString(addressString string) (*Address, error) {
	host, portString, err := net.SplitHostPort(addressString)
	if err != nil {
		return nil, err
	}

	port, err := strconv.Atoi(portString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse port: %w", err)
	}

	address := &Address{
		IP:   host,
		Port: port,
	}
	return address, nil
}

func (a *Address) String() string {
	return fmt.Sprintf("%s:%d", a.IP, a.Port)
}
