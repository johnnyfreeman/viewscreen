package style

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name         string
		disableColor bool
		wantNoColor  bool
	}{
		{
			name:         "color enabled",
			disableColor: false,
			wantNoColor:  false,
		},
		{
			name:         "color disabled",
			disableColor: true,
			wantNoColor:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Init(tt.disableColor)

			if NoColor() != tt.wantNoColor {
				t.Errorf("NoColor() = %v, want %v", NoColor(), tt.wantNoColor)
			}

			if tt.disableColor {
				if CurrentTheme != NoColorTheme {
					t.Errorf("CurrentTheme should be NoColorTheme when color is disabled")
				}
			} else {
				if CurrentTheme != DefaultTheme {
					t.Errorf("CurrentTheme should be DefaultTheme when color is enabled")
				}
			}
		})
	}
}

func TestNoColor(t *testing.T) {
	// Test when color is disabled
	Init(true)
	if !NoColor() {
		t.Errorf("NoColor() should return true after Init(true)")
	}

	// Test when color is enabled
	Init(false)
	if NoColor() {
		t.Errorf("NoColor() should return false after Init(false)")
	}
}

func TestDottedUnderline(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		disableColor bool
		wantContains string
		wantEqual    string
	}{
		{
			name:         "with color enabled",
			input:        "hello",
			disableColor: false,
			wantContains: "hello",
		},
		{
			name:         "with color disabled returns plain text",
			input:        "hello",
			disableColor: true,
			wantEqual:    "hello",
		},
		{
			name:         "empty string with color",
			input:        "",
			disableColor: false,
			wantEqual:    ansiDottedUnderline + ansiUnderlineReset,
		},
		{
			name:         "empty string without color",
			input:        "",
			disableColor: true,
			wantEqual:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Init(tt.disableColor)
			result := DottedUnderline(tt.input)

			if tt.wantEqual != "" {
				if result != tt.wantEqual {
					t.Errorf("DottedUnderline(%q) = %q, want %q", tt.input, result, tt.wantEqual)
				}
			}

			if tt.wantContains != "" {
				if !strings.Contains(result, tt.wantContains) {
					t.Errorf("DottedUnderline(%q) = %q, should contain %q", tt.input, result, tt.wantContains)
				}
			}

			// When color is enabled, verify ANSI codes are present
			if !tt.disableColor && tt.input != "" {
				if !strings.Contains(result, ansiDottedUnderline) {
					t.Errorf("DottedUnderline(%q) should contain ANSI dotted underline code", tt.input)
				}
				if !strings.Contains(result, ansiUnderlineReset) {
					t.Errorf("DottedUnderline(%q) should contain ANSI underline reset code", tt.input)
				}
			}
		})
	}
}

func TestMutedDottedUnderline(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		disableColor bool
		wantPlain    bool
	}{
		{
			name:         "with color enabled",
			input:        "/path/to/file.go",
			disableColor: false,
			wantPlain:    false,
		},
		{
			name:         "with color disabled returns plain text",
			input:        "/path/to/file.go",
			disableColor: true,
			wantPlain:    true,
		},
		{
			name:         "empty string with color",
			input:        "",
			disableColor: false,
			wantPlain:    false,
		},
		{
			name:         "empty string without color",
			input:        "",
			disableColor: true,
			wantPlain:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Init(tt.disableColor)
			result := MutedDottedUnderline(tt.input)

			if tt.wantPlain {
				if result != tt.input {
					t.Errorf("MutedDottedUnderline(%q) = %q, want plain text", tt.input, result)
				}
			} else {
				// When color is enabled, verify ANSI codes are present
				if tt.input != "" && result == tt.input {
					t.Errorf("MutedDottedUnderline(%q) should contain ANSI codes", tt.input)
				}
				// Should contain the input text
				if tt.input != "" && !strings.Contains(result, tt.input) {
					t.Errorf("MutedDottedUnderline(%q) = %q, should contain input text", tt.input, result)
				}
				// Should contain dotted underline SGR (4:4)
				if tt.input != "" && !strings.Contains(result, "4:4") {
					t.Errorf("MutedDottedUnderline(%q) = %q, should contain dotted underline code (4:4)", tt.input, result)
				}
				// Should contain foreground color (38;2 for RGB)
				if tt.input != "" && !strings.Contains(result, "38;2") {
					t.Errorf("MutedDottedUnderline(%q) = %q, should contain RGB foreground color code", tt.input, result)
				}
			}
		})
	}
}

func TestApplyGradient(t *testing.T) {
	tests := []struct {
		name         string
		text         string
		from         lipgloss.Color
		to           lipgloss.Color
		disableColor bool
		wantPlain    bool
		wantContains string
	}{
		{
			name:         "empty text",
			text:         "",
			from:         lipgloss.Color("#A855F7"),
			to:           lipgloss.Color("#22D3EE"),
			disableColor: false,
			wantPlain:    true,
		},
		{
			name:         "single character",
			text:         "A",
			from:         lipgloss.Color("#A855F7"),
			to:           lipgloss.Color("#22D3EE"),
			disableColor: false,
			wantContains: "A",
		},
		{
			name:         "multiple characters",
			text:         "Hello",
			from:         lipgloss.Color("#A855F7"),
			to:           lipgloss.Color("#22D3EE"),
			disableColor: false,
			wantContains: "H",
		},
		{
			name:         "with color disabled",
			text:         "Hello",
			from:         lipgloss.Color("#A855F7"),
			to:           lipgloss.Color("#22D3EE"),
			disableColor: true,
			wantPlain:    true,
		},
		{
			name:         "invalid from color",
			text:         "Test",
			from:         lipgloss.Color("invalid"),
			to:           lipgloss.Color("#22D3EE"),
			disableColor: false,
			wantPlain:    true,
		},
		{
			name:         "invalid to color",
			text:         "Test",
			from:         lipgloss.Color("#A855F7"),
			to:           lipgloss.Color("invalid"),
			disableColor: false,
			wantPlain:    true,
		},
		{
			name:         "unicode text",
			text:         "こんにちは",
			from:         lipgloss.Color("#A855F7"),
			to:           lipgloss.Color("#22D3EE"),
			disableColor: false,
			wantContains: "こ",
		},
		{
			name:         "same start and end color",
			text:         "Test",
			from:         lipgloss.Color("#A855F7"),
			to:           lipgloss.Color("#A855F7"),
			disableColor: false,
			wantContains: "T",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Init(tt.disableColor)
			result := ApplyGradient(tt.text, tt.from, tt.to)

			if tt.wantPlain {
				if result != tt.text {
					t.Errorf("ApplyGradient(%q) = %q, want plain text %q", tt.text, result, tt.text)
				}
			} else {
				if tt.wantContains != "" && !strings.Contains(result, tt.wantContains) {
					t.Errorf("ApplyGradient(%q) = %q, should contain %q", tt.text, result, tt.wantContains)
				}
				// When color is enabled and valid, the result should be different from input
				// (it should have ANSI codes)
				if tt.text != "" && result == tt.text {
					t.Errorf("ApplyGradient(%q) should apply color codes, got plain text", tt.text)
				}
			}
		})
	}
}

func TestApplyBoldGradient(t *testing.T) {
	tests := []struct {
		name         string
		text         string
		from         lipgloss.Color
		to           lipgloss.Color
		disableColor bool
		wantPlain    bool
		wantContains string
	}{
		{
			name:         "empty text",
			text:         "",
			from:         lipgloss.Color("#A855F7"),
			to:           lipgloss.Color("#22D3EE"),
			disableColor: false,
			wantPlain:    true,
		},
		{
			name:         "with valid colors",
			text:         "Hello",
			from:         lipgloss.Color("#A855F7"),
			to:           lipgloss.Color("#22D3EE"),
			disableColor: false,
			wantContains: "H",
		},
		{
			name:         "with color disabled",
			text:         "Hello",
			from:         lipgloss.Color("#A855F7"),
			to:           lipgloss.Color("#22D3EE"),
			disableColor: true,
			wantContains: "Hello",
		},
		{
			name:         "invalid from color",
			text:         "Test",
			from:         lipgloss.Color("bad"),
			to:           lipgloss.Color("#22D3EE"),
			disableColor: false,
			wantContains: "Test",
		},
		{
			name:         "invalid to color",
			text:         "Test",
			from:         lipgloss.Color("#A855F7"),
			to:           lipgloss.Color("bad"),
			disableColor: false,
			wantContains: "Test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Init(tt.disableColor)
			result := ApplyBoldGradient(tt.text, tt.from, tt.to)

			if tt.wantPlain {
				if result != tt.text {
					t.Errorf("ApplyBoldGradient(%q) = %q, want plain text %q", tt.text, result, tt.text)
				}
			} else if tt.wantContains != "" && !strings.Contains(result, tt.wantContains) {
				t.Errorf("ApplyBoldGradient(%q) = %q, should contain %q", tt.text, result, tt.wantContains)
			}
		})
	}
}

func TestApplyThemeGradient(t *testing.T) {
	Init(false)
	result := ApplyThemeGradient("Test")

	// Should contain the text
	if !strings.Contains(result, "T") {
		t.Errorf("ApplyThemeGradient(%q) should contain the text characters", "Test")
	}

	// Should have ANSI codes (not plain text)
	if result == "Test" {
		t.Errorf("ApplyThemeGradient(%q) should apply color codes", "Test")
	}
}

func TestApplyThemeBoldGradient(t *testing.T) {
	Init(false)
	result := ApplyThemeBoldGradient("Test")

	// Should contain the text
	if !strings.Contains(result, "T") {
		t.Errorf("ApplyThemeBoldGradient(%q) should contain the text characters", "Test")
	}

	// Should have ANSI codes (not plain text)
	if result == "Test" {
		t.Errorf("ApplyThemeBoldGradient(%q) should apply color codes", "Test")
	}
}

func TestApplySuccessGradient(t *testing.T) {
	Init(false)
	result := ApplySuccessGradient("Success")

	// Should contain the text
	if !strings.Contains(result, "S") {
		t.Errorf("ApplySuccessGradient(%q) should contain the text characters", "Success")
	}

	// Should have ANSI codes (not plain text)
	if result == "Success" {
		t.Errorf("ApplySuccessGradient(%q) should apply color codes", "Success")
	}
}

func TestApplyErrorGradient(t *testing.T) {
	Init(false)
	result := ApplyErrorGradient("Error")

	// Should contain the text
	if !strings.Contains(result, "E") {
		t.Errorf("ApplyErrorGradient(%q) should contain the text characters", "Error")
	}

	// Should have ANSI codes (not plain text)
	if result == "Error" {
		t.Errorf("ApplyErrorGradient(%q) should apply color codes", "Error")
	}
}

func TestSetTheme(t *testing.T) {
	// Create a custom theme
	customTheme := Theme{
		FgBase:   lipgloss.Color("#FFFFFF"),
		FgMuted:  lipgloss.Color("#888888"),
		FgSubtle: lipgloss.Color("#444444"),
	}

	SetTheme(customTheme)

	if CurrentTheme.FgBase != customTheme.FgBase {
		t.Errorf("SetTheme did not update CurrentTheme.FgBase")
	}
	if CurrentTheme.FgMuted != customTheme.FgMuted {
		t.Errorf("SetTheme did not update CurrentTheme.FgMuted")
	}

	// Reset to default
	SetTheme(DefaultTheme)
	if CurrentTheme != DefaultTheme {
		t.Errorf("SetTheme did not reset to DefaultTheme")
	}
}

func TestDefaultThemeColors(t *testing.T) {
	// Verify that DefaultTheme has non-empty color values
	if DefaultTheme.FgBase == "" {
		t.Error("DefaultTheme.FgBase should not be empty")
	}
	if DefaultTheme.GradientStart == "" {
		t.Error("DefaultTheme.GradientStart should not be empty")
	}
	if DefaultTheme.GradientEnd == "" {
		t.Error("DefaultTheme.GradientEnd should not be empty")
	}
	if DefaultTheme.Success == "" {
		t.Error("DefaultTheme.Success should not be empty")
	}
	if DefaultTheme.Error == "" {
		t.Error("DefaultTheme.Error should not be empty")
	}
}

func TestNoColorThemeColors(t *testing.T) {
	// Verify that NoColorTheme has all empty color values
	if NoColorTheme.FgBase != "" {
		t.Error("NoColorTheme.FgBase should be empty")
	}
	if NoColorTheme.GradientStart != "" {
		t.Error("NoColorTheme.GradientStart should be empty")
	}
	if NoColorTheme.GradientEnd != "" {
		t.Error("NoColorTheme.GradientEnd should be empty")
	}
	if NoColorTheme.Success != "" {
		t.Error("NoColorTheme.Success should be empty")
	}
	if NoColorTheme.Error != "" {
		t.Error("NoColorTheme.Error should be empty")
	}
}

func TestStyleConstants(t *testing.T) {
	// Test that style constants are defined correctly
	if Bullet != "● " {
		t.Errorf("Bullet = %q, want %q", Bullet, "● ")
	}
	if OutputPrefix != "  ⎿  " {
		t.Errorf("OutputPrefix = %q, want %q", OutputPrefix, "  ⎿  ")
	}
	if OutputContinue != "     " {
		t.Errorf("OutputContinue = %q, want %q", OutputContinue, "     ")
	}
}

func TestInitStylesColorEnabled(t *testing.T) {
	Init(false)

	// Verify that semantic styles are initialized (not zero values)
	// We can test this by rendering text and checking it's modified
	testText := "test"

	// Error style should produce colored output
	errorResult := Error.Render(testText)
	if errorResult == testText {
		t.Error("Error style should apply formatting when color is enabled")
	}

	// Success style should produce colored output
	successResult := Success.Render(testText)
	if successResult == testText {
		t.Error("Success style should apply formatting when color is enabled")
	}
}

func TestInitStylesColorDisabled(t *testing.T) {
	Init(true)

	// When color is disabled, styles should be no-ops
	testText := "test"

	// Error style should return plain text
	errorResult := Error.Render(testText)
	if errorResult != testText {
		t.Errorf("Error style should be a no-op when color is disabled, got %q", errorResult)
	}

	// Success style should return plain text
	successResult := Success.Render(testText)
	if successResult != testText {
		t.Errorf("Success style should be a no-op when color is disabled, got %q", successResult)
	}

	// Bold style should also be a no-op
	boldResult := Bold.Render(testText)
	if boldResult != testText {
		t.Errorf("Bold style should be a no-op when color is disabled, got %q", boldResult)
	}
}
