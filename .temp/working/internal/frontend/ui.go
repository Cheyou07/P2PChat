package frontend

import (
	"fmt"
	"io"
	"os"
	"unicode/utf8"

	"github.com/ShreevathsaGP/ChatP2P/internal/chat"
	"github.com/ShreevathsaGP/ChatP2P/internal/networking"

	"github.com/gdamore/tcell/v2"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/rivo/tview"
)

func BuildUI(app *tview.Application) (*tview.InputField, *tview.Form, *tview.Form, *tview.TextView, *tview.TextView, *tview.Pages) {
	//----------------UI---------------------------------------------------------------------------------------------------------------------
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
	// msgBox.SetTitle(fmt.Sprintf("Room: %s", "sample-room-name"))

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
		// msgBox.Clear()
		// pages.SwitchToPage("RoomInput") 
		// app.SetFocus(roomForm.GetFormItemByLabel("Room Name:"))
	})
	backButton.SetBorder(true).SetBackgroundColor(tcell.ColorBlack)

	peerFlex := tview.NewFlex().SetDirection(tview.FlexRow).AddItem(peersList, 0, 1, false).AddItem(backButton, 3, 0, false)
	chatPanel := tview.NewFlex().AddItem(peerFlex, peersListWidth, 1, false).AddItem(msgBox, 0, 1, false)

	// final flexboxes
	roomInputFlex := roomFormVerticalFlex
	loginFlex := loginFormVerticalFlex
	flex := tview.NewFlex().SetDirection(tview.FlexRow).AddItem(chatPanel, 0, 1, false).AddItem(input, 3, 1, true)

	pages.AddPage("RoomInput", roomInputFlex, true, true)
	pages.AddPage("Login", loginFlex, true, true)
	pages.AddPage("Chat", flex, true, false)

	app.SetRoot(pages, true).EnableMouse(true)
	app.SetFocus(loginForm.GetFormItemByLabel("First Name:"))

	return input, loginForm, roomForm, msgBox, peersList, pages
}

// Function to clear all form items
func ClearForm(form *tview.Form) {
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

func RefreshPeers(peersList *tview.TextView, cr *networking.ChatRoom, app *tview.Application) {
	if cr == nil { return }

	peers := cr.GetPeerList()
	peersList.Clear() // thead sage clear

	for _, p := range peers {
		fmt.Fprintln(peersList, ShortenID(p))
	}

	app.Draw()
}

func DisplayIncomingMessage(cm *chat.Message, msgWriter io.Writer) {
	if (cm == nil) { return }

	prompt := PrintWithColour("green", fmt.Sprintf("<%s>:", cm.SenderName))
	fmt.Fprintf(msgWriter, "%s %s\n", prompt, cm.Message)
}

func DisplayOutgoingMessage(msg string, selfName string, msgWriter io.Writer) {
	prompt := PrintWithColour("yellow", fmt.Sprintf("<%s>:", selfName))
	fmt.Fprintf(msgWriter, "%s %s\n", prompt, msg)

}

func ShortenID(p peer.ID) string {
	fullString := p.String()
	return fullString[len(fullString)-8:]
}

func PrintError(m string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, m, args...)
}

func PrintWithColour(color, msg string) string {
	return fmt.Sprintf("[%s]%s[-]", color, msg)
}