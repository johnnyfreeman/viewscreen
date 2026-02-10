package tui

// RawLineMsg is sent when a line is read from stdin
type RawLineMsg struct {
	Line string
}

// StdinClosedMsg is sent when stdin is closed
type StdinClosedMsg struct {
	Err error
}

// AutoExitTickMsg is sent each second during the auto-exit countdown.
type AutoExitTickMsg struct{}

// RerunMsg triggers a fresh claude run with the given prompt.
type RerunMsg struct {
	Prompt string
}
