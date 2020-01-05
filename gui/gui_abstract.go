package gui

import (
	"github.com/tron_client/types"
)

type ChatGui interface {
	SetChatHistory(msgs []string)
	FetchOne() (string, error)
	Close()
}

type Position struct {
	X int
	Y int
}

type PlayerKey string

const (
	Up    PlayerKey = "up"
	Down  PlayerKey = "down"
	Left  PlayerKey = "left"
	Right PlayerKey = "right"
	Key_w PlayerKey = "w"
	Key_a PlayerKey = "a"
	Key_s PlayerKey = "s"
	Key_d PlayerKey = "d"
)

type PlayerBlock struct {
	Pos   Position
	Color types.PlayerColor
}

type GameGui interface {
	SetBlocks([]PlayerBlock) error
	AppendBlocks([]PlayerBlock) error
	UserInput() PlayerKey
	Close()
	SetWin(name string)
}
