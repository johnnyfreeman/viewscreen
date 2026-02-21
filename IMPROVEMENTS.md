# Recommended Improvements

An analysis of the viewscreen codebase with prioritized suggestions organized by
impact and effort.

---

## 1. Eliminate Global Mutable State in `config` and `style`

**Impact: High | Effort: Medium**

The `config` package uses package-level `var` globals (`Verbose`, `NoColor`,
`ShowUsage`, `NoTUI`) that are written by `ParseFlags()` and read from scattered
call sites via `config.DefaultProvider{}`. The `style` package has the same
pattern with `CurrentTheme`, `noColor`, `DiffAddBg`, etc.

This creates several problems:

- **Test isolation is fragile.** Tests that call `style.Init()` or toggle
  `config.Verbose` affect every other test running in the same process. The
  `config_test.go` file already works around this by saving/restoring globals.
- **Implicit coupling.** `NewRenderer()` constructors across 5+ packages
  silently read `config.DefaultProvider{}` and `style.CurrentTheme` at
  construction time. The dependency is invisible.
- **No concurrent safety.** Nothing prevents a data race between
  `ParseFlags()` writing globals and renderer construction reading them on a
  different goroutine.

**Recommendation:** Thread a `Config` struct (the one already defined in
`config/config.go:47`) through the application from `main()`. Pass it
explicitly to `NewRendererSet()`, which distributes it to sub-renderers.
Remove `ParseFlags()` in favor of the existing `Parse()` that returns a
`*Config`. The `style.Init()` call can remain as a one-time side effect
during startup, but individual renderers should read color preferences from
the injected config rather than from globals.

This is partially started — the `Provider` interface and `Parse()` function
exist but aren't used end-to-end. The legacy `ParseFlags()` path is still the
default.

---

## 2. Use an `io.Reader` for the TUI's stdin Instead of `*bufio.Scanner`

**Impact: Medium | Effort: Low**

`tui.Model` stores a `*bufio.Scanner` as a field, which means the entire Model
struct (a Bubbletea value type) carries a reference to a stateful, non-copyable
scanner. This is awkward because Bubbletea expects `Model` to be treated as a
value — `Update()` returns a new `Model` by value.

More importantly, the scanner is created in `NewModel()` with a hardcoded
`os.Stdin` reference (`tui/model.go:50-53`), which makes the TUI model
untestable in isolation. There's no way to inject a fake stdin for testing
the `handleRawLine` → `processEvent` pipeline without rewriting the model.

**Recommendation:** Accept an `io.Reader` as a parameter to `NewModel()` (or
via a functional option). Create the scanner internally. This small change
makes the full TUI update loop testable by injecting a `bytes.Reader` or
`io.Pipe`. Consider also extracting the `ReadStdinLine` command factory to
accept the reader as a closure parameter rather than the scanner.

---

## 3. `SidebarRenderer` Is Recreated Every Render Cycle

**Impact: Medium | Effort: Low**

`RenderSidebar()` (`tui/sidebar.go:164`) creates a new `SidebarRenderer` with
`NewSidebarRenderer()` on every call. Since `View()` is called on every
Bubbletea tick (including spinner ticks, which happen ~12 times/second), this
allocates a new `SidebarRenderer`, `LogoRenderer`, and `TodoRenderer` every
frame.

```go
func RenderSidebar(s *state.State, spinner spinner.Model, height int, styles SidebarStyles) string {
    r := NewSidebarRenderer(styles, spinner) // allocation every frame
    return r.Render(s, height)
}
```

The same issue exists in `RenderHeader()` (`tui/sidebar.go:189`) which
allocates a `LogoRenderer` every frame, and in `RenderDetailsModal()` which
allocates both a `SidebarRenderer` and its children.

**Recommendation:** Store the `SidebarRenderer` as a field on `Model` and
update it when the spinner or styles change (which is rare), not on every
render. The `LogoRenderer` output is static and could be computed once.

---

## 4. Event Parsing Deserializes JSON Twice

**Impact: Medium | Effort: Low**

`events.Parse()` (`events/events.go:64-113`) first unmarshals the raw JSON into
a `BaseEvent` to read the `type` field, then unmarshals the same bytes again
into the type-specific event struct. For high-frequency `stream_event` messages
(which fire on every content delta during streaming), this doubles the JSON
parsing cost.

**Recommendation:** Use `json.RawMessage` or `json.Decoder` to extract just the
`type` field, or unmarshal into a struct that embeds `BaseEvent` plus a
`json.RawMessage` for the rest, then do the second unmarshal only for the
type-specific fields. Alternatively, since `BaseEvent` is embedded in all event
types, you could unmarshal directly into the largest type and check the `Type`
field, though this is less clean.

A simpler approach: since the JSON always has `"type":"..."` near the
beginning, use `bytes.Contains` or a small prefix check to determine the type
before any JSON parsing, then unmarshal once into the correct type.

---

## 5. Content String Builder Grows Without Bound

**Impact: High | Effort: Medium**

`tui.Model.content` is a `*strings.Builder` that accumulates all rendered
content for the viewport (`tui/model.go:28`). Every processed event appends to
it, and it's never reset or truncated. For long-running sessions with heavy tool
use (common when Claude Code runs many bash commands), this will grow
indefinitely.

Additionally, `updateViewportWithPendingTools()` (`tui/model.go:161-169`)
concatenates the full `m.content.String()` with pending tool output on every
spinner tick. Since `strings.Builder.String()` returns a copy of the internal
buffer, this means copying the entire accumulated content ~12 times per second.

**Recommendation:** Consider a ring buffer or a sliding window approach that
keeps only the last N lines (or N bytes) of rendered content. Alternatively,
implement a lazy viewport that only renders visible content. At minimum, avoid
the repeated `String()` call in the spinner tick path by caching the base
content and only appending the pending tools overlay.

---

## 6. `Truncate()` Operates on Bytes, Not Runes

**Impact: Low | Effort: Low**

`textutil.Truncate()` (`textutil/textutil.go:14-23`) uses `len(s)` and slice
indexing, which operate on bytes not Unicode code points. If the input contains
multi-byte UTF-8 characters (common in non-English tool output), the truncation
could split a multi-byte sequence, producing invalid UTF-8.

```go
func Truncate(s string, maxLen int) string {
    s = strings.TrimSpace(s)
    if len(s) <= maxLen {  // byte length, not rune length
        return s
    }
    ...
    return s[:maxLen-3] + "..."  // could split a rune
}
```

**Recommendation:** Use `[]rune(s)` or `utf8.RuneCountInString(s)` for length
checks, and convert to runes before slicing. The performance difference is
negligible for the short strings this function handles (display text, tool
arguments).

---

## 7. `WrapText()` Has Quadratic Behavior

**Impact: Low | Effort: Low**

`textutil.WrapText()` (`textutil/textutil.go:46-87`) calls
`strings.Count(result.String(), "\n")` inside the word loop to check if the
3-line limit is reached. `result.String()` copies the builder's contents, and
`strings.Count` scans the full string — both on every iteration. For a long
prompt string, this is O(n²).

**Recommendation:** Track the newline count in a local variable, incrementing
it when a newline is written. This also eliminates the `result.String()` copy
inside the loop.

---

## 8. The `docs/` Directory Has No Content

**Impact: Low | Effort: Low**

The `docs/` directory exists but its contents (if any) don't appear to be
meaningful project documentation. The README covers basic usage but doesn't
document:

- The event protocol (JSON schema for each event type)
- Architecture overview (how packages relate)
- How to add support for new tool types
- How the TUI vs. legacy streaming mode works

**Recommendation:** Add a `docs/architecture.md` covering the event flow
(stdin → parser → events → processor → renderers → viewport/stdout) and how
to extend the tool registry. This is especially valuable for an open-source
project where contributors need to understand the rendering pipeline.

---

## 9. Add a `Makefile` or `justfile` for Common Tasks

**Impact: Low | Effort: Low**

There's no build automation file. Common operations like `go build`,
`go test ./...`, `go vet ./...`, and `go test -race ./...` need to be
remembered or looked up.

**Recommendation:** Add a minimal `Makefile` with targets for `build`, `test`,
`vet`, `lint`, and `coverage`. This also provides a natural place to add CI
integration later.

---

## 10. Add Race Detection to the Test Suite

**Impact: Medium | Effort: Low**

The `indicator.Spinner` type uses a `sync.Mutex` to protect its state
(`indicator/indicator.go:22`), which suggests concurrent access is expected.
However, the test suite doesn't appear to run with `-race`. The global mutable
state in `config` and `style` packages is also a race candidate.

**Recommendation:** Run `go test -race ./...` as part of CI/development
workflow. This will surface any existing data races, particularly around the
global state discussed in recommendation #1.

---

## Summary Table

| # | Area | Impact | Effort | Category |
|---|------|--------|--------|----------|
| 1 | Global mutable state | High | Medium | Architecture |
| 2 | TUI stdin injection | Medium | Low | Testability |
| 3 | SidebarRenderer allocation | Medium | Low | Performance |
| 4 | Double JSON deserialization | Medium | Low | Performance |
| 5 | Unbounded content builder | High | Medium | Performance |
| 6 | Byte-based truncation | Low | Low | Correctness |
| 7 | Quadratic WrapText | Low | Low | Performance |
| 8 | Missing docs | Low | Low | Documentation |
| 9 | Build automation | Low | Low | Developer experience |
| 10 | Race detection | Medium | Low | Correctness |
