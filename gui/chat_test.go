package gui


import (
    "testing"
    "github.com/tron_client/client"
)

func TestChatCommunicationWithServer(t *testing.T) {
    chat := NewChat()
    stop_chan := make(chan bool)
    go client.HostServerTest(stop_chan)
    chat.Listen()
    defer chat.Close()
} 
