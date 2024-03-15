package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ShreevathsaGP/ChatP2P/internal/frontend"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

// constants
const DiscoveryInterval = time.Hour
const DiscoveryServiceTag = "chat-for-the-bois"
const RoomBufferSize = 256

func main() {
	ctx := context.Background()

	//----------------NETWORKING------------------------------------------------
	// create a new host that listens on random TCP port
	h, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"))
	if err != nil { panic(err) }

	// new pubsub using gossipsub router
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil { panic(err) }

	// local mDNS discovery
	s := mdns.NewMdnsService(h, DiscoveryServiceTag, &discoveryNotifee{h: h})
	err = s.Start()
	if err != nil { panic(err) }

	//----------------NETWORKING------------------------------------------------

	//----------------UI--------------------------------------------------------
	ui_state := frontend.BuildUI(ctx, ps, h)
	ui_state.Start(ctx)
	//----------------UI--------------------------------------------------------
}

//----------------PEER DISCOVERY----------------------------------------------
// notified on new discovery
type discoveryNotifee struct {
	h host.Host
}

// connect to peer discovered using mDNS.
// pubsub interaction begins if supported
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	// fmt.Printf("discovered new peer %s\n", pi.ID)
	err := n.h.Connect(context.Background(), pi)
	if err != nil {
		fmt.Printf("error connecting to peer %s: %s\n", pi.ID, err)
	}
}
//----------------PEER DISCOVERY----------------------------------------------