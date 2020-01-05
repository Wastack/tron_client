package engine

import (
	"fmt"
	"github.com/tron_client/client"
	"github.com/tron_client/gui"
	"github.com/tron_client/types"
	"log"
	"time"
)

type GameHandler interface {
	ListenInput()
	Close()
}

type NetGameHandler struct {
	netw   *client.Client
	engine *Game

	stopNet chan bool
}

func NewNetGameHandler(e *Game, n *client.Client) *NetGameHandler {
	return &NetGameHandler{
		netw:    n,
		engine:  e,
		stopNet: make(chan bool),
	}
}

func (h *NetGameHandler) ListenInput() {
	go func(stop chan bool) {
		for {
			select {
			case <-stop:
				log.Printf("Game phase: Listening to server stopped")
				return
			case m := <-h.netw.Msgs:
				switch m.GetType() {
				case "server_tick":
					h.processTick(m.(*types.TickMsg))
				case "error":
					log.Fatalf("Handling error message is not implemented") // TODO
				}
			}
		}
	}(h.stopNet)
	h.listenUserInput()
}

func (h *NetGameHandler) Close() {
	h.stopNet <- true
}

func (h *NetGameHandler) processTick(t *types.TickMsg) error {
	// assert last tick is correct
	if t.LastTick {
		player_alive_count := 0
		for i := range h.engine.players {
			if !h.engine.players[i].isDead {
				player_alive_count++
			}
		}
		if player_alive_count > 1 {
			return fmt.Errorf("Last tick happened with %d number of players alive",
				player_alive_count)
		}
	}
	// apply changes
	for _, change := range t.Changes {
		p, err := h.engine.playerByColor(change.Color)
		if err != nil {
			return err
		}
		// TODO verify it is not opposite direction
		if change.Dir != "" {
			p.dir = change.Dir
		}
		if p.isDead != change.Dead {
			return fmt.Errorf("Player with color %s is Dead: %t, but server's opinion is: %t",
				p.color, p.isDead, change.Dead)
		}
	}

	// time elapsed, make a step
	h.engine.Step()
	return nil
}

func (h *NetGameHandler) listenUserInput() {
	log.Printf("Game phase: listening user input")
	for {
		key := h.engine.gameGui.UserInput()
		// TODO send player event on appropiate key pressed
		switch key {
		}
	}
}

//----------------------------------------
// Local Game Handler

type LocalGameHandler struct {
	engine       *Game
	playerQueues [2]chan types.Direction

	stopped bool
}

func NewLocalGameHandler(engine *Game) *LocalGameHandler {
	return &LocalGameHandler{
		engine: engine,
		playerQueues: [...]chan types.Direction{
			make(chan types.Direction, 10),
			make(chan types.Direction, 10)},
	}
}

func (l *LocalGameHandler) ListenInput() {
	// start ticking
	go l.ticking()

	log.Printf("Game phase: listening user input")
	for {
		key := l.engine.gameGui.UserInput()
		switch key {
		case gui.Key_a:
			l.playerQueues[1] <- types.Left
		case gui.Key_w:
			l.playerQueues[1] <- types.Up
		case gui.Key_s:
			l.playerQueues[1] <- types.Down
		case gui.Key_d:
			l.playerQueues[1] <- types.Right

		case gui.Left:
			l.playerQueues[0] <- types.Left
		case gui.Up:
			l.playerQueues[0] <- types.Up
		case gui.Down:
			l.playerQueues[0] <- types.Down
		case gui.Right:
			l.playerQueues[0] <- types.Right
		}
		if l.stopped {
			log.Printf("Local game: stop receiving user input")
		}
	}
}

func (l *LocalGameHandler) Close() {
	l.stopped = true
}

func (l *LocalGameHandler) ticking() {
	for {
		// sleep will ensure ticking time
		time.Sleep(450 * time.Millisecond)

		// get one direction from each player
		// the order of playerQueues is the same as the order of players in the
		// engine
		for i, queue := range l.playerQueues {
			select {
			case dirChange := <-queue:
				// check if it is a valid direction
				// TODO pull a new element if not?
				if dirChange.Opposite() == l.engine.players[i].dir {
					break
				}
				l.engine.players[i].dir = dirChange
			default:
				// queue is empty, nothing to do here.
			}
		}
		l.engine.Step()
		if l.stopped {
			log.Printf("Local game: stop ticking")
			return
		}
	}
}
