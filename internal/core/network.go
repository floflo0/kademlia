package kademlia

type Network struct {
}

type FindNodeResponse struct {
	ipAddress string
	port      string
	ID        *KademliaID
}

func Listen(ip string, port int) {
	// TODO
}

func (network *Network) SendPingMessage(contact *Contact) {
	// TODO
}

func (network *Network) SendFindContactMessage(contact *Contact) FindNodeResponse {
	return FindNodeResponse{
		"1.1.1.1",
		"8080",
		NewKademliaID("Bonjour"),
	}
}

func (network *Network) SendFindDataMessage(hash string) {
	// TODO
}

func (network *Network) SendStoreMessage(data []byte) {
	// TODO
}
