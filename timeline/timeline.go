// Package timeline defines provider-neutral session timeline data.
package timeline

// Entry is a committed scrollback item.
type Entry struct {
	ID       string
	ParentID string
	Agent    string
	Kind     string
	Title    string
	Arg      string
	Body     string
	Lines    []string
	Nested   bool
	Status   string
}

// Text returns the rendered terminal text for the entry.
func (e Entry) Text() string {
	if e.Body != "" {
		return e.Body
	}
	if len(e.Lines) == 0 {
		return ""
	}
	n := 0
	for _, line := range e.Lines {
		n += len(line)
	}
	out := make([]byte, 0, n+len(e.Lines))
	for _, line := range e.Lines {
		out = append(out, line...)
		out = append(out, '\n')
	}
	return string(out)
}

// Activity is a live/pending timeline item.
type Activity struct {
	ID             string
	ParentID       string
	Name           string
	Input          string
	Nested         bool
	HeaderRendered bool
}

// Todo is a provider-neutral task item.
type Todo struct {
	Content    string
	Status     string
	ActiveForm string
}

// StatePatch is a declarative update to session state.
type StatePatch struct {
	Agent          *string
	Model          *string
	Version        *string
	CWD            *string
	ToolsCount     *int
	Agents         []string
	PermissionMode *string
	Prompt         *string

	IncrementTurns int
	TurnCount      *int
	TotalCost      *float64

	Todos        []Todo
	ReplaceTodos bool

	CurrentActivity *Activity
	ClearActivity   bool

	InputTokens     *int
	OutputTokens    *int
	CacheCreated    *int
	CacheRead       *int
	ReasoningTokens *int
	AddUsage        *Usage

	IsError       *bool
	DurationMS    *int
	DurationAPIMS *int
}

// Usage is an incremental token usage update.
type Usage struct {
	InputTokens     int
	OutputTokens    int
	CacheCreated    int
	CacheRead       int
	ReasoningTokens int
}

// Batch is the result of processing one provider event.
type Batch struct {
	Entries []Entry
	Patch   StatePatch
}

// StringPtr returns a pointer to s.
func StringPtr(s string) *string { return &s }

// IntPtr returns a pointer to i.
func IntPtr(i int) *int { return &i }

// FloatPtr returns a pointer to f.
func FloatPtr(f float64) *float64 { return &f }

// BoolPtr returns a pointer to b.
func BoolPtr(b bool) *bool { return &b }
