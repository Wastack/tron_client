package client

import (
    "net"
    "log"
    "github.com/tron_client/types"
    "encoding/json"
    "bufio"
    "time"
)

func HostServerTest(stop chan bool) {
    l, err := net.Listen("tcp4", ":8765")
    defer l.Close()
    c, err := l.Accept()
    if err != nil {
        log.Fatalf("Unable to connect to client")
    }

    // read connection request but do nothing with it
    _, err = bufio.NewReader(c).ReadString('\n')
    if err != nil {
	log.Printf("Unable to read connection request message")
    }

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
    time.Sleep(100 * time.Millisecond)
    // let's say Kek sends a message
    outBytes, err = json.Marshal(&types.ChatMsg{
	&types.JsonMsg{Type: "chat"},
	"Hey, what's up? I'm looking forward to play Tron with you",
	"#0000FF",
    })
    if err != nil {
        log.Fatalf("Cannot marshal chat message")
    }
    c.Write(outBytes)

    time.Sleep(100 * time.Millisecond)
    // let's say Kek sent ready
    outBytes, err = json.Marshal(&types.ReadyMsg{
	&types.JsonMsg{Type: "chat"},
	true,
	"#0000FF",
    })
    if err != nil {
        log.Fatalf("Cannot marshal chat message")
    }
    c.Write(outBytes)

}
