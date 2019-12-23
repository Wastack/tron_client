package gui

import (
	gc "github.com/rthornton128/goncurses"
	"log"
)

type NCurse struct {
	scr       *gc.Window
	outputWin *gc.Window
	inputWin  *gc.Window

	stopped chan bool
}

func NewNCurse() *NCurse {
	log.Print("Create chat window")
	screen, err := gc.Init()
	if err != nil {
		log.Fatal("Init screen:", err)
	}
	rows, cols := screen.MaxYX()
	outwin, err := gc.NewWindow(rows-3, cols-2, 0, 0)
	if err != nil {
		log.Fatal("Init output window:", err)
	}
	inwin, err := gc.NewWindow(3, cols-2, rows-3, 0)
	if err != nil {
		log.Fatal("Init input window:", err)
	}
	n := &NCurse{
		outputWin: outwin,
		inputWin:  inwin,
		scr:       screen,
		stopped:   make(chan bool),
	}

	return n
}

func (n *NCurse) clearInput() {
	n.inputWin.Erase()
	n.inputWin.Move(1, 1)
	n.inputWin.Box(gc.ACS_VLINE, gc.ACS_HLINE)
	n.inputWin.Print("> ")
	n.inputWin.NoutRefresh()
}

// ChatGui methods
// -----------------------------------

func (n *NCurse) SetChatHistory(msgs []string) {
	n.outputWin.Erase()

	// get number of available columns
	h, _ := n.outputWin.MaxYX()
	max_cols := h - 2
	start_index := 0
	if overflow := len(msgs) - max_cols; overflow > 0 {
		start_index = overflow
	}

	_ = start_index
	for i, v := range msgs[start_index:] {
		n.outputWin.Move(i+1, 1)
		n.outputWin.Println(v)
	}
	n.outputWin.Box(gc.ACS_VLINE, gc.ACS_HLINE)
	n.outputWin.NoutRefresh()
	gc.Update()
}

func (n *NCurse) Close() {
	n.outputWin.Delete()
	n.inputWin.Delete()
	gc.End()
	<-n.stopped
}

func (n *NCurse) FetchOne() (string, error) {
	for {
		n.clearInput()
		gc.Update()
		_, width := n.inputWin.MaxYX()
		inp, err := n.inputWin.GetString(width - 6)
		if err != nil {
			n.stopped <- true
			return "", err
		}
		if len(inp) < 1 {
			continue
		}
		n.clearInput()
		return inp, nil
	}
}

type HeadlessChat struct {
	Input chan string

	stop chan bool
}

func NewHeadlessChat() *HeadlessChat {
	return &HeadlessChat{
		Input: make(chan string, 1),
		stop:  make(chan bool),
	}
}

func (g *HeadlessChat) FetchOne() (string, error) {
	select {
	case msg := <-g.Input:
		return msg, nil
	case <-g.stop:
		return "", nil
	}
}

func (g *HeadlessChat) Close() {
	g.stop <- true
}

func (n *HeadlessChat) SetChatHistory(msgs []string) {}
