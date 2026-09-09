package ports

type Address struct {
	IP   string
	Port int
}

type Network interface {
	Listen(address Address) (ListenConnection, error)
	Dial(address Address) (DialConnection, error)
}

type ListenConnection interface {
	SendTo(address Address, payload []byte) error
	Receive() ([]byte, *Address, error)
	Close() error
}

type DialConnection interface {
	Send(payload []byte) error
	Receive(timeoutMiliseconds uint32) ([]byte, error)
	Close() error
}
