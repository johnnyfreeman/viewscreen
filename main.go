package main

import (
	"fmt"
	"os"

	"github.com/jfreeman/viewscreen/config"
	"github.com/jfreeman/viewscreen/parser"
)

func main() {
	config.ParseFlags()

	p := parser.NewParser()
	if err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
