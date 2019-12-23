package engine

import (
	"encoding/json"
	"fmt"
	"github.com/tron_client/client"
	"github.com/tron_client/gui"
	"github.com/tron_client/types"
	"log"
	"strconv"
	"strings"
)

const sys_n string = "Sys"

type command struct {
	description string
	execute     func(*LobbyEngine, ...string)
}

type commandMap map[string]command

var commands commandMap = commandMap{
	"/help":    {"Show help", executeHelp},
	"/connect": {"Connect to server: /connect [address] [port]", executeConnect},
	"/con":     {"", executeConnect},
	"/disc":    {"Disconnect from server", executeDisconnect},
	"/players": {"List players", executePlayers},
	"/setname": {"Set your name", executeSetname},
	"/exit":    {"Close application", func(*LobbyEngine, ...string) {}}, // exit is handle elsewhere
}

func executeSetname(c *LobbyEngine, args ...string) {
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

func executePlayers(c *LobbyEngine, _ ...string) {
	if c.net == nil {
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

func executeDisconnect(c *LobbyEngine, _ ...string) {
	if c.net == nil {
		c.PushMessage(sys_n, "You are not connected")
		return
	}
	c.stopRec <- true
	c.net.Close()
	c.net = nil
}

func (c *LobbyEngine) Close() {
	// close network connection
	if c.net != nil {
		log.Printf("Closing connection")
		c.stopRec <- true
		c.net.Close()
		c.net = nil
	}
	// close GUI
	log.Printf("Closing GUI")
	c.chatGui.Close()
}

func executeHelp(c *LobbyEngine, _ ...string) {
	c.PushMessage(sys_n, "You need help? SUCKER!! (not implemented)")
}

func executeConnect(c *LobbyEngine, args ...string) {
	if c.net != nil {
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
	resp, err := c.net.ConnectRequest(c.myPlayer.Name, "", "private")
	if err != nil {
		c.PushMessage(sys_n, fmt.Sprintf("Server error: %s", err.Error()))
	}
	c.players = resp.Players
	c.myPlayer.Color = resp.Color

	// start listening to lobby messages
	go cli.Listen()

	// Receive messages
	log.Print("Start receiving messages from server")
	go func(stop chan bool) {
		for {
			select {
			case <-stop:
				log.Printf("Listening to server stopped")
				return
			case m := <-cli.Msgs:
				log.Printf("Message received: %s", m)
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
				case "start_game":
					// advance to game phase
					c.Close()
					return
				}
			}
		}
	}(c.stopRec)

	// notify user of successfull connection
	c.PushMessage(sys_n, "Successfully connected")
}

type LobbyEngine struct {
	IsListening chan bool

	players     []types.LobbyPlayer
	myPlayer    types.LobbyPlayer
	msg_history []string

	chatGui gui.ChatGui

	net     *client.Client
	stopRec chan bool
}

func NewLobbyEngine(guiType types.GuiKind) *LobbyEngine {
	var g gui.ChatGui
	switch guiType {
	case types.NCurses:
		g = gui.NewNCurse()
	case types.Headless:
		g = gui.NewHeadlessChat()
	}

	c := LobbyEngine{
		IsListening: make(chan bool, 1),
		stopRec:     make(chan bool, 1),
		msg_history: make([]string, 0, 20),
		chatGui:     g,
	}
	c.myPlayer.Name = "Buddy"
	c.PushMessage(sys_n, "Hello! Good luck today. type '/help' for available commands")
	return &c
}

func (c *LobbyEngine) PushMessage(sender string, msg string) {
	if len(msg) < 1 {
		log.Printf("Attempt tp push empty message.")
	}
	c.msg_history = append(c.msg_history, fmt.Sprintf("%s: %s", sender, msg))
	c.chatGui.SetChatHistory(c.msg_history)
}

func (c *LobbyEngine) ListenUserInput() {
	log.Print("Start fetching messages from chat")
	for {
		var msg string
		msg, _ = c.chatGui.FetchOne()
		if msg == "" {
			log.Printf("Chat GUI closed")
			c.chatGui = nil
			return
		}

		// it's a command
		if msg[0] == '/' {

			words := strings.Fields(msg)
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
			c.PushMessage(c.myPlayer.Name, msg)
			chatMsg := &types.ChatMsg{
				JsonMsg: &types.JsonMsg{Type: "chat"},
				Message: msg,
				Color:   c.myPlayer.Color,
			}
			bytes, err := json.Marshal(chatMsg)
			if err != nil {
				log.Fatalf("Unable to marshal chat message")
			}
			if c.net != nil {
				c.net.SendMessage(bytes)
			}
		}
	}
}

func (c *LobbyEngine) playerByColor(pc types.PlayerColor) (*types.LobbyPlayer, error) {
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

func (c *LobbyEngine) removeByColor(pc types.PlayerColor) error {
	// remove from players list
	var i int
	for ; i < len(c.players); i++ {
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
