package main

import (
	"context"
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/ShreevathsaGP/ChatP2P/internal/frontend"
	"github.com/ShreevathsaGP/ChatP2P/internal/networking"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

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

// discoveryNotifee gets notified when we find a new peer via mDNS discovery
type discoveryNotifee struct {
	h host.Host
}

// HandlePeerFound connects to peers discovered via mDNS. Once they're connected,
// the PubSub system will automatically start interacting with them if they also
// support PubSub.
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	// fmt.Printf("discovered new peer %s\n", pi.ID)
	err := n.h.Connect(context.Background(), pi)
	if err != nil {
		fmt.Printf("error connecting to peer %s: %s\n", pi.ID, err)
	}
}

func main() {
	ctx := context.Background()

	//----------------NETWORKING-------------------------------------------------------------------------------------------------------------
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

	var chat_room *networking.ChatRoom = nil
	//----------------NETWORKING-------------------------------------------------------------------------------------------------------------


	//----------------UI---------------------------------------------------------------------------------------------------------------------
	app := tview.NewApplication()

	input, loginForm, roomForm, msgBox, peersList, pages := frontend.BuildUI(app)

	roomForm = roomForm.
		AddButton("Join", func() {
			roomName := roomForm.GetFormItemByLabel("Room Name:").(*tview.InputField).GetText()

			// room name length range(3, 20)
			if utf8.RuneCountInString(roomName) < 3 || utf8.RuneCountInString(roomName) > 20 {
				return
			}

			firstName := loginForm.GetFormItemByLabel("First Name:").(*tview.InputField).GetText()

			chat_room, err = networking.JoinCR(ctx, ps, h.ID(), firstName, roomName)
			if err != nil { panic(err) }

			msgBox.SetTitle(fmt.Sprintf("ROOM: %s", roomName))
			pages.SwitchToPage("Chat")
		}).AddButton("Exit", func() { 
			app.Stop()
		}).SetButtonTextColor(tcell.ColorLightYellow)

	//----------------UI---------------------------------------------------------------------------------------------------------------------


	//----------------INPUT HANDLING---------------------------------------------------------------------------------------------------------
	inputCh := make(chan string, 32)
	
	// on key enter
	input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {

			if (chat_room != nil) {
				chat_room.Leave()
			}

			msgBox.Clear()
			pages.SwitchToPage("RoomInput") 
			app.SetFocus(roomForm.GetFormItemByLabel("Room Name:"))
			return
		}

		if key != tcell.KeyEnter { return } // ignore if not enter

		line := input.GetText()
		if len(line) == 0 { return } // ignore if empty

		// exit input
		if line == "/quit" {
			app.Stop()
			return
		}

		inputCh <- line
		input.SetText("") // reset input
	})
	//----------------INPUT HANDLING---------------------------------------------------------------------------------------------------------

	
	//----------------EVENT LOOP-------------------------------------------------------------------------------------------------------------
	// buffered channel (size = 1)
	doneCh := make(chan struct{}, 1)

	go func() {
		refreshPeersTicker := time.NewTicker(time.Second)
		defer refreshPeersTicker.Stop()

		for {

			if (chat_room == nil) { continue }
			currentPage, _ := pages.GetFrontPage()
			if (currentPage != "Chat") { continue }

			select {
				case input := <- inputCh:
					err := chat_room.Publish(input)
					if err != nil { frontend.PrintError("publish error: %s", err) }

					firstName := loginForm.GetFormItemByLabel("First Name:").(*tview.InputField).GetText()

					frontend.DisplayOutgoingMessage(input, firstName, msgBox)

				case m := <- chat_room.Messages:
					frontend.DisplayIncomingMessage(m, msgBox)
				
				case <- refreshPeersTicker.C:
					frontend.RefreshPeers(peersList, chat_room, app)

				case <- ctx.Done():
					return

				case <- doneCh:
					doneCh <- struct{}{}
			}
		}
	}()
	//----------------EVENT LOOP-------------------------------------------------------------------------------------------------------------

	app.Run()
}

