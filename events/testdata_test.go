package events

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/johnnyfreeman/viewscreen/state"
)

// TestProcessAllTestdata verifies that every .jsonl file in testdata/ can be
// parsed and processed through the full EventProcessor pipeline without errors.
// This catches regressions when the Claude Code output format evolves.
func TestProcessAllTestdata(t *testing.T) {
	testdataDir := filepath.Join("..", "testdata")
	entries, err := os.ReadDir(testdataDir)
	if err != nil {
		t.Fatalf("failed to read testdata directory: %v", err)
	}

	var jsonlFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".jsonl") {
			jsonlFiles = append(jsonlFiles, entry.Name())
		}
	}

	if len(jsonlFiles) == 0 {
		t.Fatal("no .jsonl files found in testdata/")
	}

	for _, filename := range jsonlFiles {
		t.Run(filename, func(t *testing.T) {
			filePath := filepath.Join(testdataDir, filename)
			f, err := os.Open(filePath)
			if err != nil {
				t.Fatalf("failed to open %s: %v", filename, err)
			}
			defer f.Close()

			s := state.NewState()
			processor := NewEventProcessor(s)

			scanner := bufio.NewScanner(f)
			const maxCapacity = 10 * 1024 * 1024
			buf := make([]byte, maxCapacity)
			scanner.Buffer(buf, maxCapacity)

			lineNum := 0
			var parseErrors []string
			var eventCounts = make(map[string]int)

			for scanner.Scan() {
				lineNum++
				line := scanner.Text()
				if line == "" {
					continue
				}

				parsed := Parse(line)
				if parsed == nil {
					continue
				}

				switch e := parsed.(type) {
				case ParseError:
					if e.Err != nil {
						parseErrors = append(parseErrors, e.Err.Error())
					} else if !strings.HasPrefix(e.Line, "Unknown event type:") {
						parseErrors = append(parseErrors, e.Line)
					}
					// Unknown event types are logged but not fatal
					continue
				case SystemEvent:
					eventCounts["system"]++
				case SubAgentSystemEvent:
					eventCounts["subagent_system"]++
				case AssistantEvent:
					eventCounts["assistant"]++
				case UserEvent:
					eventCounts["user"]++
				case StreamEvent:
					eventCounts["stream"]++
				case ResultEvent:
					eventCounts["result"]++
				case IgnoredEvent:
					eventCounts["ignored"]++
					continue
				}

				// Process through EventProcessor - should not panic
				processor.Process(parsed)
			}

			if err := scanner.Err(); err != nil {
				t.Fatalf("scanner error: %v", err)
			}

			if len(parseErrors) > 0 {
				t.Errorf("parse errors in %s: %v", filename, parseErrors)
			}

			// Every testdata file should have at least one parseable event
			totalEvents := 0
			for _, count := range eventCounts {
				totalEvents += count
			}
			if totalEvents == 0 {
				t.Errorf("%s: no events were parsed (file has %d lines)", filename, lineNum)
			}

			t.Logf("%s: %d lines, events=%v", filename, lineNum, eventCounts)
		})
	}
}

// TestProcessTestdata_SystemEventHasModel verifies that testdata system events
// contain key fields that the current Claude Code output includes.
func TestProcessTestdata_SystemEventHasModel(t *testing.T) {
	testdataDir := filepath.Join("..", "testdata")
	entries, err := os.ReadDir(testdataDir)
	if err != nil {
		t.Fatalf("failed to read testdata directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			filePath := filepath.Join(testdataDir, entry.Name())
			f, err := os.Open(filePath)
			if err != nil {
				t.Fatalf("failed to open: %v", err)
			}
			defer f.Close()

			scanner := bufio.NewScanner(f)
			const maxCapacity = 10 * 1024 * 1024
			buf := make([]byte, maxCapacity)
			scanner.Buffer(buf, maxCapacity)

			for scanner.Scan() {
				line := scanner.Text()
				if line == "" {
					continue
				}

				parsed := Parse(line)
				sysEvent, ok := parsed.(SystemEvent)
				if !ok {
					continue
				}

				// Every system event should have a model
				if sysEvent.Data.Model == "" {
					t.Errorf("system event missing model field")
				}

				// Every system event should have a version
				if sysEvent.Data.ClaudeCodeVersion == "" {
					t.Errorf("system event missing claude_code_version field")
				}

				// Current format should include fast_mode_state
				if sysEvent.Data.FastModeState == "" {
					t.Errorf("system event missing fast_mode_state field (current format requires it)")
				}

				break // Only check the first system event
			}
		})
	}
}

// TestProcessTestdata_ResultEventHasNewFields verifies that testdata result events
// contain the new fields added in recent Claude Code versions.
func TestProcessTestdata_ResultEventHasNewFields(t *testing.T) {
	testdataDir := filepath.Join("..", "testdata")
	entries, err := os.ReadDir(testdataDir)
	if err != nil {
		t.Fatalf("failed to read testdata directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			filePath := filepath.Join(testdataDir, entry.Name())
			f, err := os.Open(filePath)
			if err != nil {
				t.Fatalf("failed to open: %v", err)
			}
			defer f.Close()

			scanner := bufio.NewScanner(f)
			const maxCapacity = 10 * 1024 * 1024
			buf := make([]byte, maxCapacity)
			scanner.Buffer(buf, maxCapacity)

			foundResult := false
			for scanner.Scan() {
				line := scanner.Text()
				if line == "" {
					continue
				}

				parsed := Parse(line)
				resultEvent, ok := parsed.(ResultEvent)
				if !ok {
					continue
				}

				foundResult = true

				// Result events should have stop_reason (added in newer versions)
				if resultEvent.Data.StopReason == "" {
					t.Errorf("result event missing stop_reason field")
				}

				// Result events should have fast_mode_state
				if resultEvent.Data.FastModeState == "" {
					t.Errorf("result event missing fast_mode_state field")
				}
			}

			if !foundResult {
				t.Skipf("no result event found in %s", entry.Name())
			}
		})
	}
}
