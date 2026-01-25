package main

import (
	"fmt"
	"os"

	"github.com/johnnyfreeman/viewscreen/config"
	"github.com/johnnyfreeman/viewscreen/parser"
)

func main() {
	config.ParseFlags()

	p := parser.NewParser()
	if err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
