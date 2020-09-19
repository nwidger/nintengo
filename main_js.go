// +build js

package main

import (
	"bytes"
	"fmt"
	"os"
	"syscall/js"

	"github.com/nwidger/nintengo/nes"
)

func main() {
	options := &nes.Options{
		Region: "NTSC",
	}

	document := js.Global().Get("document")

	canvas := document.Call("querySelector", "canvas")
	canvas.Get("style").Set("height", "50%")

	inputDiv := document.Call("createElement", "div")
	inputDiv.Get("style").Set("text-align", "center")
	document.Get("body").Call("appendChild", inputDiv)

	inputElem := document.Call("createElement", "input")
	inputElem.Call("setAttribute", "type", "file")
	inputDiv.Call("appendChild", inputElem)

	filec := make(chan js.Value, 1)
	onchangeCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go func() {
			filec <- inputElem.Get("files").Index(0)
		}()
		return nil
	})
	defer onchangeCallback.Release()
	inputElem.Set("onchange", onchangeCallback)

	file := <-filec
	gamename := file.Get("name").String()
	reader := js.Global().Get("FileReader").New()

	bufc := make(chan []byte, 1)
	onloadendCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go func() {
			result := js.Global().Get("Uint8Array").New(reader.Get("result"))
			buf := make([]byte, result.Length())
			for i := 0; i < result.Length(); i++ {
				buf[i] = byte(result.Index(i).Int())
			}
			bufc <- buf
		}()
		return nil
	})
	defer onloadendCallback.Release()
	reader.Set("onloadend", onloadendCallback)
	reader.Call("readAsArrayBuffer", file)

	buf := <-bufc
	br := bytes.NewReader(buf)

	nes, err := nes.NewNESFromReader(gamename, br, options)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

	inputDiv.Call("remove")
	canvas.Get("style").Set("height", "100%")

	nes.Run()
}
