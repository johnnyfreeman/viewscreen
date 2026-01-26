package config

import (
	"bytes"
	"flag"
	"io"
	"testing"
)

// MockStyleInitializer is a mock implementation of StyleInitializer for testing
type MockStyleInitializer struct {
	InitCalled   bool
	DisableColor bool
}

func (m *MockStyleInitializer) Init(disableColor bool) {
	m.InitCalled = true
	m.DisableColor = disableColor
}

func TestParse_DefaultValues(t *testing.T) {
	mock := &MockStyleInitializer{}
	cfg, err := Parse(
		WithArgs([]string{}),
		WithStyleInitializer(mock),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Verbose {
		t.Error("expected Verbose to be false by default")
	}

	if cfg.NoColor {
		t.Error("expected NoColor to be false by default")
	}

	if !cfg.ShowUsage {
		t.Error("expected ShowUsage to be true by default")
	}

	if !mock.InitCalled {
		t.Error("expected style initializer to be called")
	}

	if mock.DisableColor {
		t.Error("expected DisableColor to be false")
	}
}

func TestParse_VerboseFlag(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected bool
	}{
		{
			name:     "not set",
			args:     []string{},
			expected: false,
		},
		{
			name:     "short flag",
			args:     []string{"-v"},
			expected: true,
		},
		{
			name:     "explicit true",
			args:     []string{"-v=true"},
			expected: true,
		},
		{
			name:     "explicit false",
			args:     []string{"-v=false"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockStyleInitializer{}
			cfg, err := Parse(
				WithArgs(tt.args),
				WithStyleInitializer(mock),
			)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if cfg.Verbose != tt.expected {
				t.Errorf("Verbose: got %v, want %v", cfg.Verbose, tt.expected)
			}
		})
	}
}

func TestParse_NoColorFlag(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		expectedNoColor   bool
		expectedInitColor bool
	}{
		{
			name:              "not set",
			args:              []string{},
			expectedNoColor:   false,
			expectedInitColor: false,
		},
		{
			name:              "flag set",
			args:              []string{"-no-color"},
			expectedNoColor:   true,
			expectedInitColor: true,
		},
		{
			name:              "explicit true",
			args:              []string{"-no-color=true"},
			expectedNoColor:   true,
			expectedInitColor: true,
		},
		{
			name:              "explicit false",
			args:              []string{"-no-color=false"},
			expectedNoColor:   false,
			expectedInitColor: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockStyleInitializer{}
			cfg, err := Parse(
				WithArgs(tt.args),
				WithStyleInitializer(mock),
			)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if cfg.NoColor != tt.expectedNoColor {
				t.Errorf("NoColor: got %v, want %v", cfg.NoColor, tt.expectedNoColor)
			}

			if mock.DisableColor != tt.expectedInitColor {
				t.Errorf("style.Init(disableColor): got %v, want %v", mock.DisableColor, tt.expectedInitColor)
			}
		})
	}
}

func TestParse_ShowUsageFlag(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected bool
	}{
		{
			name:     "not set - defaults to true",
			args:     []string{},
			expected: true,
		},
		{
			name:     "explicit true",
			args:     []string{"-usage=true"},
			expected: true,
		},
		{
			name:     "explicit false",
			args:     []string{"-usage=false"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockStyleInitializer{}
			cfg, err := Parse(
				WithArgs(tt.args),
				WithStyleInitializer(mock),
			)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if cfg.ShowUsage != tt.expected {
				t.Errorf("ShowUsage: got %v, want %v", cfg.ShowUsage, tt.expected)
			}
		})
	}
}

func TestParse_MultipleFlagsCombined(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		wantVerbose   bool
		wantNoColor   bool
		wantShowUsage bool
	}{
		{
			name:          "all flags set",
			args:          []string{"-v", "-no-color", "-usage=false"},
			wantVerbose:   true,
			wantNoColor:   true,
			wantShowUsage: false,
		},
		{
			name:          "verbose and no-color",
			args:          []string{"-v", "-no-color"},
			wantVerbose:   true,
			wantNoColor:   true,
			wantShowUsage: true,
		},
		{
			name:          "verbose and usage disabled",
			args:          []string{"-v", "-usage=false"},
			wantVerbose:   true,
			wantNoColor:   false,
			wantShowUsage: false,
		},
		{
			name:          "no-color and usage disabled",
			args:          []string{"-no-color", "-usage=false"},
			wantVerbose:   false,
			wantNoColor:   true,
			wantShowUsage: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockStyleInitializer{}
			cfg, err := Parse(
				WithArgs(tt.args),
				WithStyleInitializer(mock),
			)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if cfg.Verbose != tt.wantVerbose {
				t.Errorf("Verbose: got %v, want %v", cfg.Verbose, tt.wantVerbose)
			}

			if cfg.NoColor != tt.wantNoColor {
				t.Errorf("NoColor: got %v, want %v", cfg.NoColor, tt.wantNoColor)
			}

			if cfg.ShowUsage != tt.wantShowUsage {
				t.Errorf("ShowUsage: got %v, want %v", cfg.ShowUsage, tt.wantShowUsage)
			}
		})
	}
}

func TestParse_InvalidFlag(t *testing.T) {
	errBuf := &bytes.Buffer{}
	mock := &MockStyleInitializer{}
	cfg, err := Parse(
		WithArgs([]string{"-invalid-flag"}),
		WithStyleInitializer(mock),
		WithErrOutput(errBuf),
	)

	if err == nil {
		t.Error("expected error for invalid flag")
	}

	if cfg != nil {
		t.Error("expected nil config on error")
	}
}

func TestParse_InvalidFlagValue(t *testing.T) {
	errBuf := &bytes.Buffer{}
	mock := &MockStyleInitializer{}
	cfg, err := Parse(
		WithArgs([]string{"-v=notabool"}),
		WithStyleInitializer(mock),
		WithErrOutput(errBuf),
	)

	if err == nil {
		t.Error("expected error for invalid flag value")
	}

	if cfg != nil {
		t.Error("expected nil config on error")
	}
}

func TestParse_WithCustomFlagSet(t *testing.T) {
	fs := flag.NewFlagSet("custom", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	mock := &MockStyleInitializer{}
	cfg, err := Parse(
		WithArgs([]string{"-v"}),
		WithFlagSet(fs),
		WithStyleInitializer(mock),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !cfg.Verbose {
		t.Error("expected Verbose to be true")
	}
}

func TestParse_WithErrOutput(t *testing.T) {
	errBuf := &bytes.Buffer{}
	mock := &MockStyleInitializer{}

	_, err := Parse(
		WithArgs([]string{"-invalid"}),
		WithStyleInitializer(mock),
		WithErrOutput(errBuf),
	)

	if err == nil {
		t.Fatal("expected error for invalid flag")
	}

	// Error output should have been written to the buffer
	if errBuf.Len() == 0 {
		t.Error("expected error output to be written to buffer")
	}
}

func TestParse_StyleInitializerCalledWithCorrectValue(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		expectedInitColor bool
	}{
		{
			name:              "color enabled",
			args:              []string{},
			expectedInitColor: false,
		},
		{
			name:              "color disabled",
			args:              []string{"-no-color"},
			expectedInitColor: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockStyleInitializer{}
			_, err := Parse(
				WithArgs(tt.args),
				WithStyleInitializer(mock),
			)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !mock.InitCalled {
				t.Fatal("expected style initializer to be called")
			}

			if mock.DisableColor != tt.expectedInitColor {
				t.Errorf("style.Init(disableColor): got %v, want %v", mock.DisableColor, tt.expectedInitColor)
			}
		})
	}
}

func TestParse_DefaultStyleInitializer(t *testing.T) {
	// Test that DefaultStyleInitializer is used when no custom initializer is provided
	// We can't easily verify this without affecting global state, so we just verify
	// that parsing succeeds without providing a custom initializer
	cfg, err := Parse(
		WithArgs([]string{"-v", "-no-color"}),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !cfg.Verbose {
		t.Error("expected Verbose to be true")
	}

	if !cfg.NoColor {
		t.Error("expected NoColor to be true")
	}
}

func TestDefaultStyleInitializer_Init(t *testing.T) {
	// Test that DefaultStyleInitializer can be instantiated and called
	// We verify the interface is implemented correctly
	var si StyleInitializer = DefaultStyleInitializer{}

	// This should not panic
	si.Init(true)
	si.Init(false)
}

func TestWithArgs(t *testing.T) {
	args := []string{"-v", "-no-color"}
	opt := WithArgs(args)

	p := &configParser{}
	opt(p)

	if len(p.args) != len(args) {
		t.Errorf("args length: got %d, want %d", len(p.args), len(args))
	}

	for i, arg := range args {
		if p.args[i] != arg {
			t.Errorf("args[%d]: got %s, want %s", i, p.args[i], arg)
		}
	}
}

func TestWithFlagSet(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	opt := WithFlagSet(fs)

	p := &configParser{}
	opt(p)

	if p.flagSet != fs {
		t.Error("expected flagSet to be set")
	}
}

func TestWithStyleInitializer(t *testing.T) {
	mock := &MockStyleInitializer{}
	opt := WithStyleInitializer(mock)

	p := &configParser{}
	opt(p)

	if p.styleInitializer != mock {
		t.Error("expected styleInitializer to be set")
	}
}

func TestWithErrOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	opt := WithErrOutput(buf)

	p := &configParser{}
	opt(p)

	if p.errOutput != buf {
		t.Error("expected errOutput to be set")
	}
}

func TestConfig_StructFields(t *testing.T) {
	// Test that Config struct has expected fields with correct types
	cfg := Config{
		Verbose:   true,
		NoColor:   true,
		ShowUsage: false,
	}

	if !cfg.Verbose {
		t.Error("expected Verbose to be true")
	}

	if !cfg.NoColor {
		t.Error("expected NoColor to be true")
	}

	if cfg.ShowUsage {
		t.Error("expected ShowUsage to be false")
	}
}

func TestParse_EmptyArgs(t *testing.T) {
	mock := &MockStyleInitializer{}
	cfg, err := Parse(
		WithArgs([]string{}),
		WithStyleInitializer(mock),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All flags should have their default values
	if cfg.Verbose != false {
		t.Error("expected Verbose to default to false")
	}

	if cfg.NoColor != false {
		t.Error("expected NoColor to default to false")
	}

	if cfg.ShowUsage != true {
		t.Error("expected ShowUsage to default to true")
	}
}

func TestParse_NilArgs(t *testing.T) {
	mock := &MockStyleInitializer{}
	cfg, err := Parse(
		WithStyleInitializer(mock),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
}

func TestParse_HelpFlag(t *testing.T) {
	errBuf := &bytes.Buffer{}
	mock := &MockStyleInitializer{}

	// -h or -help triggers ErrHelp, which is the expected behavior
	_, err := Parse(
		WithArgs([]string{"-h"}),
		WithStyleInitializer(mock),
		WithErrOutput(errBuf),
	)

	if err != flag.ErrHelp {
		t.Errorf("expected flag.ErrHelp, got: %v", err)
	}
}

func TestParse_FlagOrderIndependent(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "v before no-color",
			args: []string{"-v", "-no-color"},
		},
		{
			name: "no-color before v",
			args: []string{"-no-color", "-v"},
		},
		{
			name: "usage in middle",
			args: []string{"-v", "-usage=false", "-no-color"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockStyleInitializer{}
			cfg, err := Parse(
				WithArgs(tt.args),
				WithStyleInitializer(mock),
			)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !cfg.Verbose {
				t.Error("expected Verbose to be true")
			}

			if !cfg.NoColor {
				t.Error("expected NoColor to be true")
			}
		})
	}
}
