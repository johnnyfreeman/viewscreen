package config

import (
	"flag"
	"io"
	"strings"

	"github.com/johnnyfreeman/viewscreen/style"
)

// Provider abstracts config access for testability.
type Provider interface {
	IsVerbose() bool
	IsVeryVerbose() bool
	GetVerboseLevel() int
	NoColor() bool
	ShowUsage() bool
}

// StyleInitializer is an interface for initializing styles
type StyleInitializer interface {
	Init(disableColor bool)
}

// DefaultStyleInitializer uses the real style package
type DefaultStyleInitializer struct{}

// Init initializes styles using the real style package
func (DefaultStyleInitializer) Init(disableColor bool) {
	style.Init(disableColor)
}

// Config holds the parsed configuration.
// It implements the Provider interface directly, so it can be passed
// anywhere a Provider is needed without an adapter.
type Config struct {
	VerboseLevel int
	DisableColor bool
	DisplayUsage bool
	NoTUI        bool
	AutoExit     bool
	Dump         bool
	PromptMode   bool
	Prompt       string
}

// IsVerbose implements Provider. True at -v or higher.
func (c *Config) IsVerbose() bool { return c.VerboseLevel >= 1 }

// IsVeryVerbose implements Provider. True at -vv or higher.
func (c *Config) IsVeryVerbose() bool { return c.VerboseLevel >= 2 }

// GetVerboseLevel implements Provider.
func (c *Config) GetVerboseLevel() int { return c.VerboseLevel }

// NoColor implements Provider.
func (c *Config) NoColor() bool { return c.DisableColor }

// ShowUsage implements Provider.
func (c *Config) ShowUsage() bool { return c.DisplayUsage }

// cfg is the package-level config set by Parse().
// Accessed via Get(). This replaces the old scattered global variables.
var cfg = &Config{DisplayUsage: true}

// Get returns the current global config. Never nil.
func Get() *Config { return cfg }

// Option is a functional option for configuring the parser
type Option func(*configParser)

type configParser struct {
	flagSet          *flag.FlagSet
	args             []string
	styleInitializer StyleInitializer
	errOutput        io.Writer
}

// WithArgs sets the command line arguments to parse
func WithArgs(args []string) Option {
	return func(p *configParser) {
		p.args = args
	}
}

// WithFlagSet sets a custom flag set for parsing
func WithFlagSet(fs *flag.FlagSet) Option {
	return func(p *configParser) {
		p.flagSet = fs
	}
}

// WithStyleInitializer sets a custom style initializer
func WithStyleInitializer(si StyleInitializer) Option {
	return func(p *configParser) {
		p.styleInitializer = si
	}
}

// WithErrOutput sets the error output writer
func WithErrOutput(w io.Writer) Option {
	return func(p *configParser) {
		p.errOutput = w
	}
}

// Parse parses the provided arguments and returns a Config.
// It also sets the package-level config accessible via Get().
func Parse(opts ...Option) (*Config, error) {
	p := &configParser{
		styleInitializer: DefaultStyleInitializer{},
	}

	for _, opt := range opts {
		opt(p)
	}

	// Create a new flag set if not provided
	if p.flagSet == nil {
		p.flagSet = flag.NewFlagSet("viewscreen", flag.ContinueOnError)
	}

	if p.errOutput != nil {
		p.flagSet.SetOutput(p.errOutput)
	}

	c := &Config{
		DisplayUsage: true, // Default value
	}

	var verbose, veryVerbose, maxVerbose bool
	p.flagSet.BoolVar(&verbose, "v", false, "Verbose output (expand writes, truncated)")
	p.flagSet.BoolVar(&veryVerbose, "vv", false, "Very verbose (expand reads truncated, writes untruncated)")
	p.flagSet.BoolVar(&maxVerbose, "vvv", false, "Max verbose (expand reads with more lines)")
	p.flagSet.BoolVar(&c.DisableColor, "no-color", false, "Disable colored output")
	p.flagSet.BoolVar(&c.DisplayUsage, "usage", true, "Show token usage in result")
	p.flagSet.BoolVar(&c.NoTUI, "no-tui", false, "Disable TUI mode (use legacy streaming output)")
	p.flagSet.BoolVar(&c.AutoExit, "auto-exit", false, "Auto-exit after stream ends (useful in loops)")
	p.flagSet.BoolVar(&c.Dump, "dump", false, "Print content to stdout on TUI exit (preserves output in scrollback)")
	p.flagSet.BoolVar(&c.PromptMode, "p", false, "Treat stdin as a prompt (not a JSON stream)")

	if err := p.flagSet.Parse(p.args); err != nil {
		return nil, err
	}

	// Compute verbose level from flags (highest wins)
	if maxVerbose {
		c.VerboseLevel = 3
	} else if veryVerbose {
		c.VerboseLevel = 2
	} else if verbose {
		c.VerboseLevel = 1
	}

	// Capture positional args as prompt text
	if args := p.flagSet.Args(); len(args) > 0 {
		c.Prompt = strings.Join(args, " ")
	}

	// Auto-enable dump when auto-exit is set (loop-friendly default)
	if c.AutoExit && !isFlagSet(p.flagSet, "dump") {
		c.Dump = true
	}

	p.styleInitializer.Init(c.DisableColor)

	// Set the package-level config
	cfg = c

	return c, nil
}

// isFlagSet checks if a flag was explicitly set on the command line.
func isFlagSet(fs *flag.FlagSet, name string) bool {
	found := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
