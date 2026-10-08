package kademlia

import (
	"errors"
	"kademlia/internal/core/entities"
	"kademlia/proto/generated"
	"log/slog"

	"google.golang.org/protobuf/proto"
)

var ErrTooMuchData = errors.New("Too much data, do not exceed 1024 bytes")

func (k *kademlia) SendStore(
	recipient Contact,
	key *entities.KademliaID,
	data string,
) error {
	slog.Debug("Sending Store", "recipient ID", recipient.ID)
	connection, errDial := k.network.Dial(recipient.Address)
	if errDial != nil {
		slog.Debug("Dial")
		return errDial
	}
	defer connection.Close()

	if len([]byte(data)) > 1024 {
		return ErrTooMuchData
	}

	storeMessage := &generated.Message{
		Contact: &generated.Contact{
			KademliaId: k.me.ID[:],
			Ip:         k.me.Address.IP,
			Port:       int32(k.me.Address.Port),
		},
		Payload: &generated.Message_Store{
			Store: &generated.Store{
				Key:         key[:],
				Data:        []byte(data),
				DataLen:     int32(len(data)),
				RequesterId: k.me.ID[:],
				RecipientId: recipient.ID[:],
			},
		},
	}

	out, errMarshal := proto.Marshal(storeMessage)
	if errMarshal != nil {
		slog.Debug("Marshal")
		return errMarshal
	}

	errSend := connection.Send(out)
	if errSend != nil {
		slog.Debug("Send")
		return errSend
	}
	slog.Info("Store sent to", "recipient", recipient)

	return nil
}
