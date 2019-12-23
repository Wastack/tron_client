package main

import (
	"github.com/tron_client/engine"
	"github.com/tron_client/types"
	"log"
	"os"
)

func main() {
	f, err := os.OpenFile("tron.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	lobby := engine.NewLobbyEngine(types.NCurses)
	lobby.ListenUserInput()
	lobby.Close()
}
