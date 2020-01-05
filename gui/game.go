package gui

import (
	"fmt"
	gc "github.com/rthornton128/goncurses"
	"github.com/tron_client/types"
	"log"
)

var player_tokens = [...]byte{
	'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
	'a', 'b', 'c', 'd',
}

type NCurseGame struct {
	scr         *gc.Window
	gameWin     *gc.Window
	colors      map[types.PlayerColor]gc.Char
	token_index int
}

func NewNCurseGame(width int, height int) *NCurseGame {
	log.Print("Create game window")
	screen, err := gc.Init()
	if err != nil {
		log.Fatal("Init screen:", err)
	}
	if err := gc.StartColor(); err != nil {
		log.Fatal(err)
	}
	// need 2 characters to draw borders
	gameWin, err := gc.NewWindow(height+2, width+2, 0, 0)
	if err != nil {
		log.Fatal("Init output window:", err)
	}
	n := &NCurseGame{
		scr:     screen,
		gameWin: gameWin,
	}
	n.reset()
	gc.Update()
	return n
}

func (n *NCurseGame) reset() {
	n.gameWin.Erase()
	n.gameWin.Box(gc.ACS_VLINE, gc.ACS_HLINE)
	n.gameWin.NoutRefresh()
	n.token_index = 0
}

func (n *NCurseGame) SetBlocks(blocks []PlayerBlock) error {
	// start from scratch
	n.reset()
	return n.AppendBlocks(blocks)
}

func (n *NCurseGame) AppendBlocks(blocks []PlayerBlock) error {
	for _, b := range blocks {
		// check if color already used
		if val, ok := n.colors[b.Color]; ok {
			n.gameWin.MoveAddChar(b.Pos.Y, b.Pos.X, val)
		} else {
			if n.token_index >= len(player_tokens) {
				return fmt.Errorf("Running out of player tokens")
			}
			// assign new token
			n.colors[b.Color] = gc.Char(player_tokens[n.token_index])
			n.token_index++
			n.gameWin.MoveAddChar(b.Pos.Y, b.Pos.X, n.colors[b.Color])
		}
	}
	gc.Update()
	return nil
}

func (n *NCurseGame) Close() {
	n.gameWin.Delete()
	gc.End()
}

func (n *NCurseGame) SetWin(name string) {
	n.reset()

	// move cursor to center
	h, w := n.gameWin.MaxYX()
	n.gameWin.Move(h/2, w/2-6)

	log.Printf("SetWin called, winner is: %s", name)
	n.gameWin.Printf("Winner is: %s", name)
}

func (n *NCurseGame) UserInput() PlayerKey {
	for {
		key := n.gameWin.GetChar()
		switch key {
		case gc.KEY_UP:
			return Up
		case gc.KEY_DOWN:
			return Down
		case gc.KEY_LEFT:
			return Left
		case gc.KEY_RIGHT:
			return Right
		case 119: // w
			return Key_w
		case 97: // a
			return Key_a
		case 115: // s
			return Key_s
		case 100: // d
			return Key_d
		}
	}

}
