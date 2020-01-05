package engine

import (
	"fmt"
	"github.com/tron_client/client"
	"github.com/tron_client/gui"
	"github.com/tron_client/types"
	"log"
)

type Size struct {
	width  int
	height int
}

type playerData struct {
	history []gui.Position
	color   types.PlayerColor
	isDead  bool
	dir     types.Direction
	name    string
}

func (p *playerData) changeDir(d types.Direction) {
	if p.dir.Opposite() == d {
		log.Printf("Trying to turn to opposite direction")

	}
	p.dir = d
}

type Game struct {
	size    Size
	players []playerData

	gameGui gui.GameGui
	handler GameHandler
}

func NewGame(w int, h int, players []playerData, guik types.GuiKind, netw *client.Client) *Game {
	var gameGui gui.GameGui
	switch guik {
	case types.NCursesGame:
		gameGui = gui.NewNCurseGame(w, h)
	case types.Headless:
		// TODO
	}
	game := &Game{
		size:    Size{width: w, height: h},
		players: players,
		gameGui: gameGui,
	}
	if netw != nil {
		game.handler = NewNetGameHandler(game, netw)
	} else {
		game.handler = NewLocalGameHandler(game)
	}

	// set initial positions on GUI
	blocks := make([]gui.PlayerBlock, 0, len(players))
	for _, p := range players {
		for _, h := range p.history {
			blocks = append(blocks, gui.PlayerBlock{
				Pos:   gui.Position{X: h.X, Y: h.Y},
				Color: p.color,
			})
		}

	}
	gameGui.AppendBlocks(blocks)

	// start listening to server and user actions
	go game.handler.ListenInput()
	return game
}

func (g *Game) Step() bool {
	new_blocks := make([]gui.PlayerBlock, 0, len(g.players))

	// holds last alive player observerd. It is needed outside the loop in
	// case there is a winner
	var e *playerData

	for i := range g.players {
		e = &g.players[i]
		if e.isDead { // dead player won't step
			continue
		}
		head := e.history[len(e.history)-1]
		new_pos := move(head, e.dir)
		if new_pos.X < 0 || new_pos.X >= g.size.width ||
			new_pos.Y < 0 || new_pos.Y >= g.size.height {
			e.isDead = true
		} else {
			e.history = append(e.history, new_pos)
			new_blocks = append(new_blocks,
				gui.PlayerBlock{
					Pos:   new_pos,
					Color: e.color,
				})
		}
	}
	if len(new_blocks) == 0 {
		g.gameGui.SetWin("") // it is draw
	} else if len(new_blocks) == 1 {
		g.gameGui.SetWin(e.name) // there is a winner
	}
	g.gameGui.AppendBlocks(new_blocks)
	return false
}

func move(pos gui.Position, dir types.Direction) gui.Position {
	switch dir {
	case types.Up:
		return gui.Position{X: pos.X, Y: pos.Y - 1}
	case types.Down:
		return gui.Position{X: pos.X, Y: pos.Y + 1}
	case types.Right:
		return gui.Position{X: pos.X + 1, Y: pos.Y}
	case types.Left:
		return gui.Position{X: pos.X - 1, Y: pos.Y}
	default:
		panic("Invalid direction")
	}
}

func (g *Game) playerByColor(c types.PlayerColor) (*playerData, error) {
	for i := range g.players {
		if c == g.players[i].color {
			return &g.players[i], nil
		}
	}
	return nil, fmt.Errorf("Unable to find player color")
}
