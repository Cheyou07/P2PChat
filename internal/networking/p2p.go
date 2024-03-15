package networking

import (
	"context"
	"encoding/json"

	"github.com/ShreevathsaGP/ChatP2P/internal/chat"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
)

// Reference: go-libp2p/examples/pubsub/chat [https://github.com/libp2p/go-libp2p/tree/master/examples]

// no. of incoming messages per topic
const RoomBufferSize = 256

type ChatRoom struct {
	Messages chan *chat.Message

	ctx   context.Context
	ps    *pubsub.PubSub
	topic *pubsub.Topic
	sub   *pubsub.Subscription

	roomName 	string
	selfID   	peer.ID
	selfName	string
}

// subscribe to a pubsub topic
func JoinCR(ctx context.Context, ps *pubsub.PubSub, selfID peer.ID, selfName string, roomName string) (*ChatRoom, error) {
	// join the topic (get ability to sub)
	topic, err := ps.Join(getTopicName(roomName))
	if err != nil {
		return nil, err
	}

	// subscribe to the topic
	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	chat_room := &ChatRoom{
		ctx:      ctx,
		ps:       ps,
		topic:    topic,
		sub:      sub,
		selfID:  	selfID,
		selfName:	selfName,
		roomName: roomName,
		Messages:	make(chan *chat.Message, RoomBufferSize),
	}

	// start reading loop
	go chat_room.readLoop()
	return chat_room, nil
}

func (chat_room *ChatRoom) Leave() {
	chat_room.sub.Cancel()
	chat_room.topic.Close()
}

// publish to pubsub topic
func (chat_room *ChatRoom) Publish(message string) error {
	m := chat.Message{
		Message:    message,
		SenderID:   chat_room.selfID.String(),
		SenderName: chat_room.selfName,
	}
	msgBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return chat_room.topic.Publish(chat_room.ctx, msgBytes)
}

func (chat_room *ChatRoom) GetName() string{
	return chat_room.selfName
}

func (chat_room *ChatRoom) GetPeerList() []peer.ID {
	return chat_room.ps.ListPeers(getTopicName(chat_room.roomName))
}

// pull messages from the pubsub topic
// put messages into Messages channel
func (chat_room *ChatRoom) readLoop() {
	for {
		msg, err := chat_room.sub.Next(chat_room.ctx)
		if err != nil {
			close(chat_room.Messages)
			return
		}

		// only forward incoming messages
		// ie. messages coming from others
		if msg.ReceivedFrom == chat_room.selfID {
			continue
		}
		cm := new(chat.Message)
		
		err = json.Unmarshal(msg.Data, cm)
		if err != nil { continue }
		
		// valid messages into channel
		chat_room.Messages <- cm
	}
}

func getTopicName(roomName string) string {
	return "chat-room:" + roomName
}