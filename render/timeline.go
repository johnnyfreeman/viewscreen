package render

import (
	"strings"

	"github.com/johnnyfreeman/viewscreen/style"
	"github.com/johnnyfreeman/viewscreen/timeline"
)

// TimelineRenderer renders provider-neutral timeline entries and activities.
type TimelineRenderer struct{}

// NewTimelineRenderer creates a timeline renderer.
func NewTimelineRenderer() *TimelineRenderer {
	return &TimelineRenderer{}
}

// RenderEntry returns the committed terminal text for an entry.
func (r *TimelineRenderer) RenderEntry(entry timeline.Entry) string {
	return entry.Text()
}

// RenderEntries returns the committed terminal text for entries.
func (r *TimelineRenderer) RenderEntries(entries []timeline.Entry) string {
	var sb strings.Builder
	for _, entry := range entries {
		sb.WriteString(r.RenderEntry(entry))
	}
	return sb.String()
}

// RenderActivity returns terminal text for a live pending item.
func (r *TimelineRenderer) RenderActivity(activity timeline.Activity, icon string) string {
	var sb strings.Builder
	if activity.Nested {
		sb.WriteString(style.OutputPrefix)
	}
	sb.WriteString(style.ApplyThemeBoldGradient(icon + " " + activity.Name))
	if activity.Input != "" {
		sb.WriteString(" " + style.MutedText(truncateTimelineArg(activity.Input, 80)))
	}
	sb.WriteString("\n")
	return sb.String()
}

func truncateTimelineArg(s string, max int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}
