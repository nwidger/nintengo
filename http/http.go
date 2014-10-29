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
	NES             *nes.NES
	PTLeft          string
	PTRight         string
	CPUMemory       string
	PPUMemory       string
	OAMMemory       string
	OAMBufferMemory string
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
			cpuMemory[i] = neserv.NES.CPU.Memory.Memory.Fetch(uint16(i))
		}

		page.CPUMemory = hex.Dump(cpuMemory)

		ppuMemory := make([]byte, m65go2.DEFAULT_MEMORY_SIZE)

		for i := uint32(0); i < m65go2.DEFAULT_MEMORY_SIZE; i++ {
			ppuMemory[i] = neserv.NES.PPU.Memory.Memory.Fetch(uint16(i))
		}

		page.PPUMemory = hex.Dump(ppuMemory)

		oamMemory := make([]byte, 256)

		for i := uint32(0); i < 256; i++ {
			oamMemory[i] = neserv.NES.PPU.OAM.BasicMemory.Fetch(uint16(i))
		}

		page.OAMMemory = hex.Dump(oamMemory)

		oamBufferMemory := make([]byte, 32)

		for i := uint32(0); i < 32; i++ {
			oamBufferMemory[i] = neserv.NES.PPU.OAM.Buffer.Fetch(uint16(i))
		}

		page.OAMBufferMemory = hex.Dump(oamBufferMemory)

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
