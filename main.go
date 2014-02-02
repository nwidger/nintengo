package main

import (
	"fmt"
	"github.com/nwidger/nintengo/nes"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: <rom-file>\n")
		return
	}

	nes, err := nes.NewNES(os.Args[1])

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

	nes.Run()
}
