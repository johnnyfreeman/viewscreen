package config

import (
	"flag"

	"github.com/jfreeman/viewscreen/style"
)

var (
	Verbose   bool
	NoColor   bool
	ShowUsage bool
)

// ParseFlags parses command line flags and configures the application
func ParseFlags() {
	flag.BoolVar(&Verbose, "v", false, "Verbose output (show more details)")
	flag.BoolVar(&NoColor, "no-color", false, "Disable colored output")
	flag.BoolVar(&ShowUsage, "usage", true, "Show token usage in result")
	flag.Parse()

	style.Init(NoColor)
}
