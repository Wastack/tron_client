package main

import (
    "github.com/tron_client/gui"
    "os"
    "log"
)

func main() {
    f, err := os.OpenFile("tron.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
    if err != nil {
	log.Fatalf("error opening file: %v", err)
    }
    defer f.Close()
    log.SetOutput(f)

    chat := gui.NewChat()
    chat.Listen()
    chat.Close()
}
