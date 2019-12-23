package gui

type ChatGui interface {
	SetChatHistory(msgs []string)
	FetchOne() (string, error)
	Close()
}
