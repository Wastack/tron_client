package types

type Message string
type PlayerColor string

type LobbyPlayer struct {
	Color PlayerColor `json:"color"`
	Name  string      `json:"name"`
	Ready bool        `json:"ready"`
}

type JsonMsgI interface {
	GetType() string
}

type JsonMsg struct {
	Type string `json:"type"`
}

type ConnReqMsg struct {
	*JsonMsg        // "connect"
	Name     string `json:"name"`
	GroupId  string `json:"id,omitempty"`
	Privacy  string `json:"privacy"`
}

// TODO setting missing
type ConnRespMsg struct {
	*JsonMsg
	Color   PlayerColor   `json:"color"`
	Players []LobbyPlayer `json:"players"`
	Id      string        `json:"id"`
}

type ReadyMsg struct {
	*JsonMsg
	Value bool        `json:"value"`
	Color PlayerColor `json:"color,omitempty"`
}

type ChatMsg struct {
	*JsonMsg
	Message string      `json:"message"`
	Color   PlayerColor `json:"color"`
}

type ConnAckMsg struct {
	*JsonMsg
	Player LobbyPlayer `json:"player"`
	Action string      `json:"action"`
}

type TickMsg struct {
	*JsonMsg
	Countdown int          `json:"countdown"`
	Changes   []GameChange `json:"changes"`
	LastTick  bool         `json:"lasttick"`
}

type GameChange struct {
	Color PlayerColor `json:"color"`
	Dir   Direction   `json:"direction"`
	Dead  bool        `json:"dead"`
}

type PlayerEventMsg struct {
	*JsonMsg
	Color PlayerColor `json:"color"`
	Dir   Direction   `json:"direction"`
}

type Direction string

func (d Direction) Opposite() Direction {
	switch d {
	case Up:
		return Down
	case Down:
		return Up
	case Left:
		return Right
	case Right:
		return Left
	default:
		panic("Unknown direction")
	}
}

const (
	Up    = "up"
	Down  = "down"
	Left  = "left"
	Right = "right"
)

func (m *JsonMsg) GetType() string {
	return m.Type
}

type GuiKind int

const (
	NCursesLobby GuiKind = iota
	NCursesGame  GuiKind = iota
	Headless     GuiKind = iota
)
