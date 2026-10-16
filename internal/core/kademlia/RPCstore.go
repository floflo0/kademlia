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
	connection, errDial := k.network.Dial(recipient.Address)
	if errDial != nil {
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
		return errMarshal
	}

	errSend := connection.Send(out)
	if errSend != nil {
		return errSend
	}
	slog.Info("Sending STORE RPC", "recipientID", recipient.ID)

	return nil
}
