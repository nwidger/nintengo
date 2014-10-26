package http

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"

	"html/template"

	"image/png"

	"encoding/hex"

	"github.com/nwidger/nintengo/m65go2"
	"github.com/nwidger/nintengo/nes"
)

type Page struct {
	NES       *nes.NES
	PTLeft    string
	PTRight   string
	CPUMemory string
	PPUMemory string
}

type NEServer struct {
	*nes.NES
	address string
}

func NewNEServer(nes *nes.NES, addr string) *NEServer {
	return &NEServer{
		NES:     nes,
		address: addr,
	}
}

func (neserv *NEServer) Run() (err error) {
	var t *template.Template

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		page := Page{
			NES: neserv.NES,
		}

		cpuMemory := make([]byte, m65go2.DEFAULT_MEMORY_SIZE)

		for i := uint32(0); i < m65go2.DEFAULT_MEMORY_SIZE; i++ {
			cpuMemory[i] = neserv.NES.CPU.Memory.Fetch(uint16(i))
		}

		page.CPUMemory = hex.Dump(cpuMemory)

		ppuMemory := make([]byte, m65go2.DEFAULT_MEMORY_SIZE)

		for i := uint32(0); i < m65go2.DEFAULT_MEMORY_SIZE; i++ {
			ppuMemory[i] = neserv.NES.PPU.Memory.Fetch(uint16(i))
		}

		page.PPUMemory = hex.Dump(ppuMemory)

		left, right := neserv.NES.PPU.GetPatternTables()

		buf := new(bytes.Buffer)
		png.Encode(buf, left)
		page.PTLeft = base64.StdEncoding.EncodeToString(buf.Bytes())

		buf = new(bytes.Buffer)
		png.Encode(buf, right)
		page.PTRight = base64.StdEncoding.EncodeToString(buf.Bytes())

		t, err = template.New("index").Parse(index)

		if err != nil {
			fmt.Printf("*** Error parsing template: %s\n", err)
			return
		}

		err = t.Execute(w, page)

		if err != nil {
			fmt.Printf("*** Error executing template: %s\n", err)
			return
		}

	})

	err = http.ListenAndServe(neserv.address, nil)

	return
}
