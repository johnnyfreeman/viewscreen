// Package render provides rendering utilities for terminal output.
package render

import (
	"io"
	"strings"
)

// Output wraps an io.Writer to provide both direct writing and string collection.
// This eliminates the need for duplicate Render() and RenderToString() methods
// across renderer packages - instead, callers use WriterOutput for direct output
// or StringOutput for string collection, and a single render method handles both.
type Output struct {
	w io.Writer
}

// WriterOutput creates an Output that writes to the given io.Writer.
func WriterOutput(w io.Writer) *Output {
	return &Output{w: w}
}

// StringOutput creates an Output that collects output into a string.
// Call String() on the returned Output to get the collected content.
func StringOutput() *Output {
	return &Output{w: &strings.Builder{}}
}

// Write implements io.Writer.
func (o *Output) Write(p []byte) (n int, err error) {
	return o.w.Write(p)
}

// WriteString writes a string to the output.
func (o *Output) WriteString(s string) (n int, err error) {
	return io.WriteString(o.w, s)
}

// String returns the collected output if this Output was created with StringOutput.
// Returns an empty string if created with WriterOutput.
func (o *Output) String() string {
	if sb, ok := o.w.(*strings.Builder); ok {
		return sb.String()
	}
	return ""
}
