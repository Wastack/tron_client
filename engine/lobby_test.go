package engine


import (
    "testing"
    "github.com/tron_client/gui"
    "github.com/tron_client/types"
    "net"
    "log"
    "encoding/json"
    "bufio"
    "time"
    "os"
    "fmt"
)

type mockServer struct {
    con net.Conn
    ready chan bool
}

func (m *mockServer) Close() {
    if m.con != nil {
	m.con.Close()
    }
}

func (m *mockServer) hostServer() {
    l, err := net.Listen("tcp4", ":8765")
    defer l.Close()
    c, err := l.Accept()
    if err != nil {
        log.Fatalf("Unable to connect to client")
    }
    m.con = c
    fmt.Println("hostServer: accepted connection successfully")

    // read connection request but do nothing with it
    _, err = bufio.NewReader(c).ReadString('\n')
    if err != nil {
	log.Printf("Unable to read connection request message")
    }
    fmt.Println("hostServer: connection request received")

    outJson := &types.ConnRespMsg {
	&types.JsonMsg{Type: "connect"},
	"#FF0000",
	[]types.LobbyPlayer{{Color: "#00FF00", Name: "Zold", Ready: true}, {Color: "#0000FF", Name: "Kek", Ready: false}},
	"dsgjngohnthgkjdflkn",
    }
    outBytes, err := json.Marshal(outJson)
    if err != nil {
        log.Fatalf("Listen: malformed chat message")
    }
    c.Write(outBytes)
    m.ready <- true

}

func assertEqual(t *testing.T, a interface{}, b interface{}) {
    if a == b {
	return
    }
    t.Fatalf("%v != %v", a, b)
}

func TestChatCommunicationWithServer(t *testing.T) {
    lobby := NewLobbyEngine(types.Headless)


    // start server
    server := mockServer{ready: make(chan bool)}
    defer server.Close()

    // server should shut down if client's disconnected successfully. No need to
    // shut down server manually.
    go server.hostServer()

    // make engine listen to GUI
    go lobby.Listen()
    defer lobby.Close()

    // assume user called /connect
    lobby.chatGui.(*gui.HeadlessChat).Input <- "/connect"

    fmt.Println("Waiting for server")
    <-time.After(1*time.Second)
    fmt.Println("done waiting")

    if len(lobby.players) != 2 {
	t.Fatalf("Number of players incorrect: %d", len(lobby.players))
    }
    assertEqual(t, lobby.players[0].Color, "#00FF00")
    assertEqual(t, lobby.players[1].Color, "#0000FF")
    assertEqual(t, lobby.players[0].Name, "Zold")
    assertEqual(t, lobby.players[1].Name, "Kek")
    assertEqual(t, lobby.players[0].Ready, true)

    //chatHistoryCount := len(lobby.msg_history)

    //// let's say Kek sent a message
    //outBytes, err := json.Marshal(&types.ChatMsg{
    //    &types.JsonMsg{Type: "chat"},
    //    "Hey, what's up? I'm looking forward to play Tron with you",
    //    "#0000FF",
    //})
    //if err != nil {
    //    log.Fatalf("Cannot marshal chat message")
    //}
    //server.con.Write(outBytes)

    //// assert for a new entry in message history
    //assertEqual(t, len(lobby.msg_history), chatHistoryCount+1)
    //// history  should contain the message
    //if !strings.Contains(lobby.msg_history[chatHistoryCount], 
    //        "Hey, what's up? I'm looking forward to play Tron with you") {
    //    t.Fatalf("Received message is not in history")
    //}

    //// let's say Kek sent ready
    //outBytes, err = json.Marshal(&types.ReadyMsg{
    //    &types.JsonMsg{Type: "chat"},
    //    true,
    //    "#0000FF",
    //})
    //if err != nil {
    //    log.Fatalf("Cannot marshal chat message")
    //}
    //server.con.Write(outBytes)
 
    //// player's state should change to true
    //assertEqual(t, lobby.players[0].Ready, true)

    //stop_chan <- true
}
