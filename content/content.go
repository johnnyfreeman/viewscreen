// Package content provides utilities for extracting text from JSON content
// structures used in Claude Code events.
package content

import (
	"encoding/json"
	"strings"
)

// Block represents a single content block in an array of content blocks.
// Claude Code events sometimes encode content as an array of typed blocks.
type Block struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ExtractText extracts text from a json.RawMessage that can be either:
// - A JSON string: "hello world"
// - An array of content blocks: [{"type": "text", "text": "hello"}]
//
// Returns the extracted text, or the raw string representation as fallback.
func ExtractText(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	// Try to unmarshal as string first
	var str string
	if err := json.Unmarshal(raw, &str); err == nil {
		return str
	}

	// Try to unmarshal as array of content blocks
	var blocks []Block
	if err := json.Unmarshal(raw, &blocks); err == nil {
		return extractTextFromBlocks(blocks)
	}

	// Fallback: return raw string representation
	return string(raw)
}

// extractTextFromBlocks extracts text from an array of content blocks,
// joining text blocks with newlines.
func extractTextFromBlocks(blocks []Block) string {
	var parts []string
	for _, block := range blocks {
		if block.Type == "text" && block.Text != "" {
			parts = append(parts, block.Text)
		}
	}
	return strings.Join(parts, "\n")
}
