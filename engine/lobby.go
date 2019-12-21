package gui

import (
    "github.com/tron_client/types"
    gc "github.com/rthornton128/goncurses"
    "github.com/tron_client/client"
    "log"
    "strings"
    "fmt"
    "strconv"
)

const sys_n string = "Sys"

type command struct {
    description string
    execute func(*Chat, ...string)
}

type commandMap map[string]command

var commands commandMap= commandMap{
    "/help" : {"Show help", executeHelp},
    "/connect" : {"Connect to server: /connect [address] [port]", executeConnect},
    "/con" : {"", executeConnect},
    "/disc": {"Disconnect from server", executeDisconnect},
    "/players": {"List players", executePlayers},
    "/setname": {"Set your name", executeSetname},
    "/exit": {"Close application", func(*Chat,...string){}}, // exit is handle elsewhere
}

func executeSetname(c *Chat, args...string) {
    if len(args) < 1 {
	c.PushMessage(sys_n, fmt.Sprintf("Your name is: %s", c.myPlayer.Name))
	return
    }
    name := &args[0]
    if len(*name) > 30 || len(*name) < 3 {
	c.PushMessage(sys_n, "Length of name should be betwwen 3 and 30")
	return
    }
    c.myPlayer.Name = *name
}

func executePlayers(c *Chat, _...string) {
    if !c.net.IsConnected() {
        c.PushMessage(sys_n, "You are not connected")
    }

    // list players including this client
    c.PushMessage(sys_n, fmt.Sprintf("Player: %s, Color: %s, Ready: %t",
        c.myPlayer.Name, c.myPlayer.Color, c.myPlayer.Ready))
    for i := range c.players {
	c.PushMessage(sys_n, fmt.Sprintf("Player: %s, Color: %s, Ready: %t",
	    c.players[i].Name, c.players[i].Color, c.players[i].Ready))
    }
}

func executeDisconnect(c *Chat, _...string) {
    if !c.net.IsConnected() {
        c.PushMessage(sys_n, "You are not connected")
	return
    }
    c.stopRec <- true
    c.net.Close()
}

func executeHelp(c *Chat, _...string) {
    c.PushMessage(sys_n, "You need help? SUCKER!! (not implemented)")
}

func executeConnect(c *Chat, args...string) {
    if c.net.IsConnected(){
        c.PushMessage(sys_n, "You are already connected. Try to disconnect first with: '/disc[onnect]")
    }
    address, port := "localhost", 8765
    if len(args) > 0 {
        address = args[1]
    }
    if len(args) > 1 {
        port_candid, err := strconv.Atoi(args[1])
        if err != nil {
	   c.PushMessage(sys_n, "Port is not a valid number.")
	   return
        }
        port = port_candid
    }
    cli, err := client.Connect(address, port)
    if err != nil {
        c.PushMessage(sys_n, "Could not connect to server")
	return
    }
    c.net = cli
    log.Print("Start listening to server")

    resp, err := c.net.ConnectRequest(c.myPlayer.Name, "", "private")
    if err != nil {
	c.PushMessage(sys_n, fmt.Sprintf("Server error: %s", err.Error()))
    }
    c.players = resp.Players
    c.myPlayer.Color = resp.Color

    // start listening to lobby messages
    go cli.Listen()

    // Receive messages
    go func(stop chan bool) {
	for {
	    select {
	    case <-stop:
	        return
	    case m := <- cli.Msgs:
		switch m.GetType() {
		case "chat":
		    chatMsg := m.(*types.ChatMsg)
		    p, err := c.playerByColor(chatMsg.Color)
	            if err != nil {
		       c.PushMessage(sys_n, "Server error")
	            }
		    c.PushMessage(p.Name, chatMsg.Message)
		case "ready":
		    r := m.(*types.ReadyMsg)
		    p, err := c.playerByColor(r.Color)
	            if err != nil {
		       c.PushMessage(sys_n, "Server error")
	            }
	            // assign new ready value
	            p.Ready = r.Value
		    c.PushMessage(sys_n, fmt.Sprintf("%s set ready to %t", p.Name, r.Value))
		case "connection":
		    ack := m.(*types.ConnAckMsg)
		    switch ack.Action {
		    case "disconnect":
			c.PushMessage(sys_n, fmt.Sprintf(
			    "Player %s (%s) disconnected", ack.Player.Name, ack.Player.Color))
			err = c.removeByColor(ack.Player.Color)
			if err != nil {
			    c.PushMessage(sys_n, "Error: player unknown")
			}
		    case "connec":
			c.PushMessage(sys_n, fmt.Sprintf(
			    "Player %s (%s) connected", ack.Player.Name, ack.Player.Color))
			// add to players list
			c.players = append(c.players, ack.Player)
		    default:
			c.PushMessage(sys_n, "Error: malformed message")
		    }
		}
	    }
	}
    }(c.stopRec)
}

type Chat struct {
    UserInput chan string

    players []types.LobbyPlayer
    myPlayer types.LobbyPlayer
    msg_history []string


    scr *gc.Window
    outputWin *gc.Window
    inputWin *gc.Window

    net *client.Client
    stopRec chan bool
}

func NewChat(guiType types.GuiKind) *Chat {
    log.Print("Create chat window")
    screen, err := gc.Init()
    if err != nil {
	log.Fatal("Init screen:", err)
    }
    rows, cols := screen.MaxYX()
    outwin, err := gc.NewWindow(rows-3, cols-2, 0, 0)
    if err != nil {
	log.Fatal("Init output window:", err)
    }
    inwin, err := gc.NewWindow(3, cols-2, rows-3, 0)
    if err != nil {
	log.Fatal("Init input window:", err)
    }
    c := Chat {
	UserInput: make(chan string, 1),
	stopRec: make(chan bool),
	scr: screen,
	outputWin: outwin,
	inputWin: inwin,
	msg_history: make([]string, 0, 20),
    }
    c.myPlayer.Name = "Buddy"

    // clear
    c.PushMessage(sys_n, "Hello! Good luck today. type '/help' for available commands")
    c.clearInput()
    gc.Update()
    return &c
}

func (c *Chat) PushMessage(sender string, msg string) {
    if len(msg) < 1 {
	log.Printf("Attempt tp push empty message.")
    }
    c.msg_history = append(c.msg_history, fmt.Sprintf("%s: %s", sender, msg))
    c.outputWin.Erase()

    // get number of available columns
    h, _ := c.outputWin.MaxYX()
    max_cols := h - 2
    start_index := 0
    if overflow := len(c.msg_history) - max_cols; overflow > 0 {
	start_index = overflow
    }

    _ = start_index
    for i, v := range c.msg_history[start_index:] {
    	c.outputWin.Move(i+1,1)
    	c.outputWin.Println(v)
    }
    c.outputWin.Box(gc.ACS_VLINE, gc.ACS_HLINE)
    c.outputWin.NoutRefresh()
    gc.Update()
}


func (c *Chat) Close() {
    c.outputWin.Delete()
    c.inputWin.Delete()
    gc.End()
}

func (c *Chat) clearInput() {
    c.inputWin.Erase()
    c.inputWin.Move(1,1)
    c.inputWin.Box(gc.ACS_VLINE, gc.ACS_HLINE)
    c.inputWin.Print("> ")
    c.inputWin.NoutRefresh()
}

func (c *Chat) Listen() {
    for {
	c.clearInput()
    	gc.Update()
	_, width := c.inputWin.MaxYX()
	inp, err := c.inputWin.GetString(width-6)
	if err != nil {
	    log.Fatal("Could not fetch user input:", err)
	}
	if len(inp) < 1 {
	    continue
	}
	c.clearInput()

	// it's a command
	if inp[0] == '/'{

	    words := strings.Fields(inp)
	    if words[0] == "/exit" {
		return
	    }
	    if command, ok := commands[words[0]]; ok {
		command.execute(c, words[1:]...)
	    } else {
		c.PushMessage(sys_n, fmt.Sprintf("Unkown command: '%s'", words[0]))
	    }
	} else {
	    // simple message
	    c.PushMessage(c.myPlayer.Name, inp)
	}
    }
}

func (c *Chat) playerByColor(pc types.PlayerColor) (*types.LobbyPlayer, error) {
    if pc == c.myPlayer.Color {
	return &c.myPlayer, nil
    }
    for i := range c.players {
	if c.players[i].Color == pc {
	    return &c.players[i], nil
	}
    }
    return nil, fmt.Errorf("Unknown player identifier")
}

func (c *Chat) removeByColor(pc types.PlayerColor) error {
    // remove from players list
    var i int
    for ; i< len(c.players); i++ {
        if c.players[i].Color == pc {
    	break
        }
    }

    // if not found
    if i >= len(c.players) {
	return fmt.Errorf("Could not find player by color")
    }
    c.players = append(c.players[:i], c.players[i+1:]...)
    return nil
}
