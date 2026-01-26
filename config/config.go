package config

import (
	"flag"
	"io"

	"github.com/johnnyfreeman/viewscreen/style"
)

var (
	Verbose   bool
	NoColor   bool
	ShowUsage bool
)

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

// Config holds the parsed configuration
type Config struct {
	Verbose   bool
	NoColor   bool
	ShowUsage bool
}

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

// Parse parses the provided arguments and returns a Config
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

	cfg := &Config{
		ShowUsage: true, // Default value
	}

	p.flagSet.BoolVar(&cfg.Verbose, "v", false, "Verbose output (show more details)")
	p.flagSet.BoolVar(&cfg.NoColor, "no-color", false, "Disable colored output")
	p.flagSet.BoolVar(&cfg.ShowUsage, "usage", true, "Show token usage in result")

	if err := p.flagSet.Parse(p.args); err != nil {
		return nil, err
	}

	p.styleInitializer.Init(cfg.NoColor)

	return cfg, nil
}

// ParseFlags parses command line flags and configures the application
// This is the legacy function that uses global variables
func ParseFlags() {
	flag.BoolVar(&Verbose, "v", false, "Verbose output (show more details)")
	flag.BoolVar(&NoColor, "no-color", false, "Disable colored output")
	flag.BoolVar(&ShowUsage, "usage", true, "Show token usage in result")
	flag.Parse()

	style.Init(NoColor)
}
