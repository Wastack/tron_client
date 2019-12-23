package engine

import (
	"bufio"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/tron_client/gui"
	"github.com/tron_client/types"
	"log"
	"net"
	"strings"
	"testing"
	"time"
)

type mockServer struct {
	con   net.Conn
	ready chan bool
}

func (m *mockServer) Close() {
	if m.con != nil {
		m.con.Close()
	}
}

func (m *mockServer) hostServer() {
	log.SetFlags(log.Lshortfile)
	l, err := net.Listen("tcp4", ":8765")
	defer l.Close()
	if err != nil {
		log.Fatalf("Unable to listen on port 8765: %s", err.Error())
	}
	c, err := l.Accept()
	if err != nil {
		log.Fatalf("Unable to connect to client")
	}
	m.con = c

	// read connection request but do nothing with it
	_, err = bufio.NewReader(c).ReadString('\n')
	if err != nil {
		log.Printf("Unable to read connection request message")
	}

	outJson := &types.ConnRespMsg{
		JsonMsg: &types.JsonMsg{Type: "connect"},
		Color:   "#FF0000",
		Players: []types.LobbyPlayer{{Color: "#00FF00", Name: "Zold", Ready: true}, {Color: "#0000FF", Name: "Kek", Ready: false}},
		Id:      "dsgjngohnthgkjdflkn",
	}
	outBytes, err := json.Marshal(outJson)
	if err != nil {
		log.Fatalf("Listen: malformed chat message")
	}
	m.sendMessage(outBytes)
	close(m.ready)

}

func (m *mockServer) sendMessage(msg []byte) {
	m.con.Write(append(msg, '\n'))
}

func TestChatCommunicationWithServer(t *testing.T) {
	assert := assert.New(t)
	lobby := NewLobbyEngine(types.Headless)

	// start server
	server := mockServer{ready: make(chan bool)}

	// server should shut down if client's disconnected successfully. No need to
	// shut down server manually.
	go server.hostServer()
	defer server.Close()

	// make engine listen to GUI
	go lobby.ListenUserInput()

	// assume user called /connect
	lobby.chatGui.(*gui.HeadlessChat).Input <- "/connect"

	select {
	case <-time.After(1 * time.Second):
		t.Fatalf("Timeout while waiting for server")
	case <-server.ready:
		t.Logf("Server ready")
	}
	// wait for connect response to be processed
	time.Sleep(20 * time.Millisecond)

	if len(lobby.players) != 2 {
		t.Fatalf("Number of players incorrect: %d", len(lobby.players))
	}
	assert.Equal(types.PlayerColor("#00FF00"), lobby.players[0].Color)
	assert.Equal(types.PlayerColor("#0000FF"), lobby.players[1].Color)
	assert.Equal("Zold", lobby.players[0].Name)
	assert.Equal("Kek", lobby.players[1].Name)
	assert.Equal(true, lobby.players[0].Ready)

	chatHistoryCount := len(lobby.msg_history)

	// let's say Kek sent a message
	outBytes, err := json.Marshal(&types.ChatMsg{
		JsonMsg: &types.JsonMsg{Type: "chat"},
		Message: "Hey, what's up? I'm looking forward to play Tron with you",
		Color:   "#0000FF",
	})
	if err != nil {
		log.Fatalf("Cannot marshal chat message")
	}
	server.sendMessage(outBytes)
	time.Sleep(20 * time.Millisecond)

	// assert for a new entry in message history
	assert.Equal(chatHistoryCount+1, len(lobby.msg_history))
	// history  should contain the message
	if !strings.Contains(lobby.msg_history[chatHistoryCount],
		"Hey, what's up? I'm looking forward to play Tron with you") {
		t.Fatalf("Received message is not in history")
	}

	// let's say Kek sent ready
	outBytes, err = json.Marshal(&types.ReadyMsg{
		JsonMsg: &types.JsonMsg{Type: "ready"},
		Value:   true,
		Color:   types.PlayerColor("#0000FF"),
	})
	if err != nil {
		log.Fatalf("Cannot marshal chat message")
	}
	server.sendMessage(outBytes)
	time.Sleep(20 * time.Millisecond)

	// player's state should change to true
	assert.Equal(lobby.players[1].Ready, true)

	// client send's ready signal
	lobby.chatGui.(*gui.HeadlessChat).Input <- "/ready"
	// server should receive ready
	msg, err := bufio.NewReader(server.con).ReadString('\n')
	assert.Nil(err)
	readyMsg := &types.ReadyMsg{}
	err = json.Unmarshal([]byte(msg), readyMsg)
	assert.True(readyMsg.Value)

	// assume server sends start game
	outBytes, _ = json.Marshal(&types.JsonMsg{Type: "start_game"})
	server.sendMessage(outBytes)

	time.Sleep(20 * time.Millisecond)
	assert.Nil(lobby.net)
	assert.Nil(lobby.chatGui)

}
