package types

// MarkdownRenderer abstracts markdown rendering for testability.
// This interface is used by assistant, stream, and user packages to allow
// dependency injection of the markdown rendering implementation.
type MarkdownRenderer interface {
	Render(content string) string
	SetWidth(width int)
}
