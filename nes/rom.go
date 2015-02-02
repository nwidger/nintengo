package nes

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"archive/zip"

	"strings"

	"github.com/nwidger/nintengo/rp2ago3"
	"github.com/nwidger/nintengo/rp2cgo2"
)

func init() {
	gob.Register(&ROMFile{})

	gob.Register(&ANROM{})
	gob.Register(&CNROM{})
	gob.Register(&MMC1{})
	gob.Register(&MMC2{})
	gob.Register(&MMC3{})
	gob.Register(&NROM{})
	gob.Register(&UNROM{})
}

//go:generate stringer -type=Region
type Region uint8

const (
	NTSC Region = iota
	PAL
)

type ROMFile struct {
	Gamename    string
	PRGBanks    uint16
	CHRBanks    uint16
	Mirroring   rp2cgo2.Mirroring
	Battery     bool
	Trainer     bool
	FourScreen  bool
	VSCart      bool
	Mapper      uint8
	RAMBanks    uint8
	RegionFlag  Region
	TrainerData []uint8
	WRAMBanks   [][]uint8
	ROMBanks    [][]uint8
	VROMBanks   [][]uint8
	irq         func(state bool)
	setTables   func(t0, t1, t2, t3 int)
}

type ROM interface {
	rp2ago3.MappableMemory
	Region() Region
	String() string
	GameName() string
	LoadBattery()
	SaveBattery() (err error)
	GetROMFile() *ROMFile
}

func getBuf(filename string) (buf []byte, suffix string, err error) {
	var r *zip.ReadCloser
	var rc io.ReadCloser

	switch {
	case strings.HasSuffix(filename, ".nes") || strings.HasSuffix(filename, ".NES"):
		suffix = filename[len(filename)-len(".nes"):]
		buf, err = ioutil.ReadFile(filename)
	case strings.HasSuffix(filename, ".zip") || strings.HasSuffix(filename, ".ZIP"):
		suffix = filename[len(filename)-len(".zip"):]

		// Open a zip archive for reading.
		r, err = zip.OpenReader(filename)

		if err != nil {
			return
		}

		defer r.Close()

		// Iterate through the files in the archive,
		// printing some of their contents.
		for _, f := range r.File {
			if !strings.HasSuffix(f.Name, ".nes") && !strings.HasSuffix(f.Name, ".NES") {
				continue
			}

			rc, err = f.Open()

			if err != nil {
				return
			}

			buf, err = ioutil.ReadAll(rc)

			if err != nil {
				return
			}

			rc.Close()
			break
		}
	default:
		err = errors.New("Unknown filetype, must be .nes, .NES, .zip or .ZIP")
	}

	return
}

func NewROMFromRaw(gamename string, raw []byte, irq func(state bool), setTables func(t0, t1, t2, t3 int)) (rom ROM, err error) {
	romf, err := NewROMFile(raw)

	if err != nil {
		return
	}

	romf.irq = irq
	romf.setTables = setTables

	romf.setTables(romf.Tables())
	romf.Gamename = gamename

	switch romf.Mapper {
	case 0x00, 0x40, 0x41:
		rom = NewNROM(romf)
	case 0x01:
		rom = NewMMC1(romf)
	case 0x02, 0x42:
		rom = NewUNROM(romf)
	case 0x03, 0x43:
		rom = NewCNROM(romf)
	case 0x04:
		rom = NewMMC3(romf)
	case 0x07:
		rom = NewANROM(romf)
	case 0x09:
		rom = NewMMC2(romf)
	default:
		err = errors.New(fmt.Sprintf("Unsupported mapper type %v", romf.Mapper))
	}

	return
}

func NewROM(filename string, irq func(state bool), setTables func(t0, t1, t2, t3 int)) (rom ROM, err error) {
	var buf []byte
	var suffix string

	buf, suffix, err = getBuf(filename)

	if err != nil {
		return
	}

	gamename := strings.TrimSuffix(filename, suffix)
	rom, err = NewROMFromRaw(gamename, buf, irq, setTables)
	return
}

func NewROMFile(buf []byte) (romf *ROMFile, err error) {
	var offset int

	if len(buf) < 16 {
		err = errors.New("Invalid ROM: Missing 16-byte header")
		return
	}

	if string(buf[0:3]) != "NES" || buf[3] != 0x1a {
		err = errors.New("Invalid ROM: Missing 'NES' constant in header")
		return
	}

	romf = &ROMFile{}

	i := 4

	for ; i < 10; i++ {
		byte := buf[i]

		switch i {
		case 4:
			romf.PRGBanks = uint16(byte)
		case 5:
			romf.CHRBanks = uint16(byte)
		case 6:
			for j := 0; j < 4; j++ {
				if byte&(0x01<<uint8(j)) != 0 {
					switch j {
					case 0:
						romf.Mirroring = rp2cgo2.Vertical
					case 1:
						romf.Battery = true
					case 2:
						romf.Trainer = true
					case 3:
						romf.FourScreen = true
						romf.Mirroring = rp2cgo2.FourScreen
					}
				}
			}

			romf.Mapper = (byte >> 4) & 0x0f
		case 7:
			if byte&0x01 != 0 {
				romf.VSCart = true
			}

			romf.Mapper |= byte & 0xf0

		case 8:
			romf.RAMBanks = byte

			if romf.RAMBanks == 0 {
				romf.RAMBanks = 1
			}
		case 9:
			if byte&0x01 != 0 {
				romf.RegionFlag = PAL
			}
		}
	}

	i += 6

	if romf.Trainer {
		offset = 512

		if len(buf) < (i + offset) {
			romf = nil
			err = errors.New("Invalid ROM: EOF in trainer data")
			return
		}

		romf.TrainerData = buf[i : i+offset]
		i += offset
	}

	offset = 1024 * 16

	if len(buf) < (i + (offset * int(romf.PRGBanks))) {
		romf = nil
		err = errors.New("Invalid ROM: EOF in ROM bank data")
		return
	}

	romf.ROMBanks = make([][]uint8, romf.PRGBanks)

	for n := 0; n < int(romf.PRGBanks); n++ {
		romf.ROMBanks[n] = buf[i : i+offset]
		i += offset
	}

	offset = 1024 * 8

	if len(buf) < (i + (offset * int(romf.CHRBanks))) {
		romf = nil
		err = errors.New("Invalid ROM: EOF in VROM bank data")
		return
	}

	romf.VROMBanks = make([][]uint8, romf.CHRBanks)

	for n := 0; n < int(romf.CHRBanks); n++ {
		romf.VROMBanks[n] = buf[i : i+offset]
		i += offset
	}

	offset = 1024 * 8

	romf.WRAMBanks = make([][]uint8, romf.RAMBanks)

	for n := 0; n < int(romf.RAMBanks); n++ {
		romf.WRAMBanks[n] = make([]uint8, offset)
	}

	return
}

func (romf *ROMFile) Region() Region {
	return romf.RegionFlag
}

func (romf *ROMFile) Tables() (t0, t1, t2, t3 int) {
	switch romf.Mirroring {
	case rp2cgo2.Horizontal:
		t0, t1, t2, t3 = 0, 0, 1, 1
	case rp2cgo2.Vertical:
		t0, t1, t2, t3 = 0, 1, 0, 1
	}

	return
}

func (romf *ROMFile) String() string {
	return fmt.Sprintf("PRG Banks: %v\n", romf.PRGBanks) +
		fmt.Sprintf("CHR Banks: %v\n", romf.CHRBanks) +
		fmt.Sprintf("Mirroring: %v\n", romf.Mirroring) +
		fmt.Sprintf("Battery: %v\n", romf.Battery) +
		fmt.Sprintf("Trainer: %v\n", romf.Trainer) +
		fmt.Sprintf("FourScreen: %v\n", romf.FourScreen) +
		fmt.Sprintf("VS Cart: %v\n", romf.VSCart) +
		fmt.Sprintf("RAM Banks: %v\n", romf.RAMBanks) +
		fmt.Sprintf("Region: %v\n", romf.RegionFlag)
}

func (romf *ROMFile) GameName() string {
	return romf.Gamename
}

func (romf *ROMFile) LoadBattery() {
	var ram []byte

	if !romf.Battery || romf.RAMBanks == 0 {
		return
	}

	savename := romf.Gamename + ".sav"
	ram, err := ioutil.ReadFile(savename)

	if err != nil {
		return
	}

	fmt.Println("*** Loading battery from " + savename)

	for b := range romf.WRAMBanks {
		for i := uint16(0); i < 0x2000; i++ {
			romf.WRAMBanks[b][i] = ram[i]
		}
	}

	return
}

func (romf *ROMFile) SaveBattery() (err error) {
	if !romf.Battery || romf.RAMBanks == 0 {
		return
	}

	savename := romf.Gamename + ".sav"

	fmt.Println("*** Saving battery to " + savename)

	buf := bytes.Buffer{}

	for b := range romf.WRAMBanks {
		for i := uint16(0); i < 0x2000; i++ {
			buf.WriteByte(romf.WRAMBanks[b][i])
		}
	}

	err = ioutil.WriteFile(savename, buf.Bytes(), 0644)

	return
}

func (romf *ROMFile) GetROMFile() *ROMFile {
	return romf
}
