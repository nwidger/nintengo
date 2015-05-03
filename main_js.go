// +build js

package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/gopherjs/gopherjs/js"
	"github.com/nwidger/nintengo/nes"
)

func main() {
	options := &nes.Options{
		Region: "NTSC",
	}

	document := js.Global.Get("document")
	inputElem := document.Call("createElement", "input")
	inputElem.Call("setAttribute", "type", "file")
	document.Get("body").Call("appendChild", inputElem)

	filec := make(chan *js.Object, 1)
	inputElem.Set("onchange", func(event *js.Object) {
		filec <- inputElem.Get("files").Index(0)
	})

	file := <-filec
	gamename := file.Get("name").String()
	reader := js.Global.Get("FileReader").New()

	bufc := make(chan []byte, 1)
	reader.Set("onloadend", func(event *js.Object) {
		bufc <- js.Global.Get("Uint8Array").New(reader.Get("result")).Interface().([]byte)
	})
	reader.Call("readAsArrayBuffer", file)

	buf := <-bufc
	br := bytes.NewReader(buf)

	nes, err := nes.NewNESFromReader(gamename, br, options)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

	inputElem.Call("remove")

	go nes.Run()
}
