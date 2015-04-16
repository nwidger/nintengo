// +build js

package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/nwidger/nintengo/nes"
)

func main() {
	options := &nes.Options{
		Region: "NTSC",
	}

	buf, err := Asset(`Super Mario Bros. (W) [!].nes`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

	br := bytes.NewReader(buf)
	nes, err := nes.NewNESFromReader(br, options)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

	_ = nes
	// go nes.Run()
}
