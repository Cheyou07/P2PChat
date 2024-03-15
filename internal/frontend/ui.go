package frontend

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"
	"unicode/utf8"

	"github.com/ShreevathsaGP/ChatP2P/internal/chat"
	"github.com/ShreevathsaGP/ChatP2P/internal/networking"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type UI struct {
	chat_room *networking.ChatRoom
	app 			*tview.Application
	peersList	*tview.TextView
	pages 		*tview.Pages

	msgWriter io.Writer

	inputCh 	chan string
	doneCh 		chan struct{}
}

func BuildUI(ctx context.Context, ps *pubsub.PubSub, h host.Host) *UI {
	app := tview.NewApplication()
	var err error
	
	// initialise UI struct
	ui_state := &UI{
		chat_room: nil,
		app: app,
		peersList: nil,
		msgWriter: nil,
		inputCh: make(chan string, 32),
		doneCh: make(chan struct{}, 1),
	}
	peersListWidth := 20
		
	// calculate dimensions
	// terminalFd := int(os.Stdout.Fd())
	// boxWidth, boxHeight := 40, 10
	// terminalWidth, terminalHeight, err := term.GetSize(terminalFd)
	// x, y := (terminalWidth - boxWidth) / 2, (terminalHeight - boxHeight) / 2

	var pages = tview.NewPages()

	// text view contains chat messages
	msgBox := tview.NewTextView()
	msgBox.SetDynamicColors(true)
	msgBox.SetBorder(true)

	// text view = io.Writers that dont automatically refresh
	// force app to redraw when new messages come
	msgBox.SetChangedFunc(func() { app.Draw() })

	// message input field
	input := tview.NewInputField().
		SetLabel(" ENTER MESSAGE" + ": ").
		SetFieldWidth(0).
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetLabelColor(tcell.ColorOrange)

	input.SetBorder(true)

	// list of peers (modified by refreshPeers)
	peersList := tview.NewTextView()
	peersList.SetBorder(true)
	peersList.SetTitle("PEERS")
	peersList.SetChangedFunc(func() { app.Draw() })

	roomForm := tview.NewForm().
		AddInputField("Room Name:", "gen", 30, nil, nil).SetLabelColor(tcell.ColorOrange)

	loginForm := tview.NewForm().
		AddInputField("First Name:", "", 20, nil, nil).SetLabelColor(tcell.ColorOrange).
		AddInputField("Last Name:", "", 20, nil, nil).SetLabelColor(tcell.ColorOrange).
		AddDropDown("Gender:", []string{
				"Male",
				"Female",
		}, 0, nil).SetFieldTextColor(tcell.ColorLightYellow).
		AddCheckbox("Under development. Still wish to proceed?", true, nil).
		AddCheckbox("First name & Last name must be 3 to 20 chars long!", false, nil)

	loginForm = loginForm.
		AddButton("Save", func() {
			firstName := loginForm.GetFormItemByLabel("First Name:").(*tview.InputField).GetText()
			lastName := loginForm.GetFormItemByLabel("Last Name:").(*tview.InputField).GetText()

			// names length min(3)
			if utf8.RuneCountInString(firstName) < 3 || utf8.RuneCountInString(lastName) < 3 {
				return
			}
			// names length max(20)
			if utf8.RuneCountInString(firstName) > 20 || utf8.RuneCountInString(lastName) > 20 {
					return
			}

			pages.SwitchToPage("RoomInput")
		}).SetButtonTextColor(tcell.ColorLightYellow).
		AddButton("Exit", func() { 
			app.Stop()
		}).SetButtonTextColor(tcell.ColorLightYellow)

	roomForm = roomForm.
	AddButton("Join", func() {
		roomName := roomForm.GetFormItemByLabel("Room Name:").(*tview.InputField).GetText()

		// room name length range(3, 20)
		if utf8.RuneCountInString(roomName) < 3 || utf8.RuneCountInString(roomName) > 20 {
			return
		}

		firstName := loginForm.GetFormItemByLabel("First Name:").(*tview.InputField).GetText()

		ui_state.chat_room, err = networking.JoinCR(ctx, ps, h.ID(), firstName, roomName)
		if err != nil { panic(err) }

		msgBox.SetTitle(fmt.Sprintf("ROOM: %s", roomName))
		pages.SwitchToPage("Chat")
	}).AddButton("Exit", func() { 
		app.Stop()
	}).SetButtonTextColor(tcell.ColorLightYellow)

	genderDropdown := loginForm.GetFormItemByLabel("Gender:").(*tview.DropDown)
	genderDropdown.SetFieldWidth(20)

	loginForm.SetBorder(true).SetTitle("LOGIN")
	roomForm.SetBorder(true).SetTitle("ROOM-SELECTION")

	// center login form horizontally
	loginFormHorizontalFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(tview.NewTextView(), 0, 1, false).
		AddItem(loginForm, 75, 1, true).
		AddItem(tview.NewTextView(), 0, 1, false)

	// center login form vertically
	loginFormVerticalFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewTextView(), 0, 1, false).
		AddItem(loginFormHorizontalFlex, 15, 1, true).
		AddItem(tview.NewTextView(), 0, 1, false)

	// center room form horizontally
	roomFormHorizontalFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(tview.NewTextView(), 0, 1, false).
		AddItem(roomForm, 45, 1, true).
		AddItem(tview.NewTextView(), 0, 1, false)

	// center room form vertically
	roomFormVerticalFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewTextView(), 0, 1, false).
		AddItem(roomFormHorizontalFlex, 7, 1, true).
		AddItem(tview.NewTextView(), 0, 1, false)

	// back button
	backButton := tview.NewButton("BACK").SetSelectedFunc(func() {
		resetForm(roomForm)
		msgBox.Clear()
		pages.SwitchToPage("RoomInput")
		app.SetFocus(roomForm.GetFormItemByLabel("Room Name:"))
	})
	backButton.SetBorder(true).SetBackgroundColor(tcell.ColorBlack)
	backButton.SetBackgroundColor(tcell.ColorBlack)

	peerFlex := tview.NewFlex().SetDirection(tview.FlexRow).AddItem(peersList, 0, 1, false).AddItem(backButton, 3, 0, false)
	chatPanel := tview.NewFlex().AddItem(peerFlex, peersListWidth, 1, false).AddItem(msgBox, 0, 1, false)

	//----------------INPUT HANDLING--------------------------------------
	// on key enter
	input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			if (ui_state.chat_room != nil) { ui_state.chat_room.Leave() }

			msgBox.Clear()
			pages.SwitchToPage("RoomInput") 
			app.SetFocus(roomForm.GetFormItemByLabel("Room Name:"))
			return
		}

		if key != tcell.KeyEnter { return }

		line := input.GetText()
		if len(line) == 0 { return } // ignore if empty

		// exit input
		if line == "/quit" {
			app.Stop()
			return
		}

		ui_state.inputCh <- line
		input.SetText("") // reset input
	})
	//----------------INPUT HANDLING--------------------------------------

	// final flexboxes
	roomInputFlex := roomFormVerticalFlex
	loginFlex := loginFormVerticalFlex
	flex := tview.NewFlex().SetDirection(tview.FlexRow).AddItem(chatPanel, 0, 1, false).AddItem(input, 3, 1, true)

	pages.AddPage("RoomInput", roomInputFlex, true, true)
	pages.AddPage("Login", loginFlex, true, true)
	pages.AddPage("Chat", flex, true, false)

	ui_state.pages = pages
	ui_state.msgWriter = msgBox
	ui_state.peersList = peersList

	app.SetRoot(pages, true).EnableMouse(true)
	app.SetFocus(loginForm.GetFormItemByLabel("First Name:"))

	return ui_state
}

func (ui_state *UI) startEventHandler(ctx context.Context) {
	refreshPeersTicker := time.NewTicker(time.Second)
	defer refreshPeersTicker.Stop()

	for {

		if (ui_state.chat_room == nil) { continue }
		currentPage, _ := ui_state.pages.GetFrontPage()
		if (currentPage != "Chat") { continue }

		select {
			case input := <- ui_state.inputCh:
				err := ui_state.chat_room.Publish(input)
				if err != nil { printError("publish error: %s", err) }

				displayOutgoingMessage(input, ui_state.chat_room.GetName(), ui_state.msgWriter)

			case m := <- ui_state.chat_room.Messages:
				displayIncomingMessage(m, ui_state.msgWriter)
			
			case <- refreshPeersTicker.C:
				refreshPeers(ui_state.peersList, ui_state.chat_room, ui_state.app)

			case <- ctx.Done():
				return

			case <- ui_state.doneCh:
				ui_state.doneCh <- struct{}{}
		}
	}
}

// start the ui
func (ui_state *UI) Start(ctx context.Context) error {
	go ui_state.startEventHandler(ctx)
	defer ui_state.end()

	return ui_state.app.Run()
}

func (ui_state *UI) end() {
	ui_state.doneCh <- struct{}{}
}

// reset all form items
func resetForm(form *tview.Form) {
	for index := 0; index < form.GetFormItemCount(); index++ {
		item := form.GetFormItem(index)
		switch item := item.(type) {
		case *tview.InputField:
			item.SetText("")
		case *tview.DropDown:
			item.SetCurrentOption(0)
		case *tview.Checkbox:
			item.SetChecked(true)
		}
	}
}

func refreshPeers(peersList *tview.TextView, cr *networking.ChatRoom, app *tview.Application) {
	if cr == nil { return }

	peers := cr.GetPeerList()
	peersList.Clear() // thead sage clear

	for _, p := range peers {
		fmt.Fprintln(peersList, shortenID(p))
	}

	app.Draw()
}

func displayIncomingMessage(cm *chat.Message, msgWriter io.Writer) {
	if (cm == nil) { return }

	prompt := printWithColour("green", fmt.Sprintf("<%s>:", cm.SenderName))
	fmt.Fprintf(msgWriter, "%s %s\n", prompt, cm.Message)
}

func displayOutgoingMessage(msg string, selfName string, msgWriter io.Writer) {
	prompt := printWithColour("yellow", fmt.Sprintf("<%s>:", selfName))
	fmt.Fprintf(msgWriter, "%s %s\n", prompt, msg)

}

func shortenID(p peer.ID) string {
	fullString := p.String()
	return fullString[len(fullString)-10:]
}

func printError(m string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, m, args...)
}

func printWithColour(color, msg string) string {
	return fmt.Sprintf("[%s]%s[-]", color, msg)
}