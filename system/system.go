package system

import (
	"fmt"
	"strings"

	"github.com/jfreeman/viewscreen/config"
	"github.com/jfreeman/viewscreen/style"
	"github.com/jfreeman/viewscreen/types"
)

// Event represents a system initialization event
type Event struct {
	types.BaseEvent
	Subtype           string   `json:"subtype"`
	CWD               string   `json:"cwd"`
	Tools             []string `json:"tools"`
	Model             string   `json:"model"`
	PermissionMode    string   `json:"permissionMode"`
	ClaudeCodeVersion string   `json:"claude_code_version"`
	Agents            []string `json:"agents"`
}

// Render outputs the system event to the terminal
func Render(event Event) {
	// Use gradient for session header when color is enabled
	header := fmt.Sprintf("%sSession Started", style.Bullet)
	if !style.NoColor() {
		header = style.ApplyThemeBoldGradient(header)
	} else {
		header = style.SessionHeader.Render(header)
	}
	fmt.Println(header)
	fmt.Printf("%s%s %s\n", style.OutputPrefix, style.Muted.Render("Model:"), event.Model)
	fmt.Printf("%s%s %s\n", style.OutputContinue, style.Muted.Render("Version:"), event.ClaudeCodeVersion)
	fmt.Printf("%s%s %s\n", style.OutputContinue, style.Muted.Render("CWD:"), event.CWD)
	fmt.Printf("%s%s %d available\n", style.OutputContinue, style.Muted.Render("Tools:"), len(event.Tools))
	if config.Verbose && len(event.Agents) > 0 {
		fmt.Printf("%s%s %s\n", style.OutputContinue, style.Muted.Render("Agents:"), strings.Join(event.Agents, ", "))
	}
	fmt.Println()
}
