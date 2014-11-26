package nes

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"archive/zip"

	"strings"

	"github.com/nwidger/nintengo/rp2ago3"
	"github.com/nwidger/nintengo/rp2cgo2"
)

type Region uint8

func (r Region) String() string {
	switch r {
	case NTSC:
		return "NTSC"
	case PAL:
		return "PAL"
	}

	return "Unknown"
}

const (
	NTSC Region = iota
	PAL
)

type ROMFile struct {
	gamename    string
	filename    string
	prgBanks    uint16
	chrBanks    uint16
	mirroring   rp2cgo2.Mirroring
	battery     bool
	trainer     bool
	fourScreen  bool
	vsCart      bool
	mapper      uint8
	ramBanks    uint8
	region      Region
	trainerData []uint8
	wramBanks   [][]uint8
	romBanks    [][]uint8
	vromBanks   [][]uint8
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
}

func getBuf(filename string) (buf []byte, suffix string, err error) {
	var r *zip.ReadCloser
	var rc io.ReadCloser

	switch {
	case strings.HasSuffix(filename, ".nes"):
		suffix = ".nes"

		if strings.HasSuffix(filename, suffix) {
			buf, err = ioutil.ReadFile(filename)
			return
		}
	case strings.HasSuffix(filename, ".zip"):
		suffix = ".zip"

		// Open a zip archive for reading.
		r, err = zip.OpenReader(filename)

		if err != nil {
			return
		}

		defer r.Close()

		// Iterate through the files in the archive,
		// printing some of their contents.
		for _, f := range r.File {
			if !strings.HasSuffix(f.Name, ".nes") {
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
		err = errors.New("Unknown filetype, must be .nes or .zip")
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

	romf, err := NewROMFile(buf)

	if err != nil {
		return
	}

	romf.irq = irq
	romf.setTables = setTables
	romf.filename = filename
	romf.gamename = strings.TrimSuffix(romf.filename, suffix)

	romf.setTables(romf.Tables())

	switch romf.mapper {
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
		err = errors.New(fmt.Sprintf("Unsupported mapper type %v", romf.mapper))
	}

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
			romf.prgBanks = uint16(byte)
		case 5:
			romf.chrBanks = uint16(byte)
		case 6:
			for j := 0; j < 4; j++ {
				if byte&(0x01<<uint8(j)) != 0 {
					switch j {
					case 0:
						romf.mirroring = rp2cgo2.Vertical
					case 1:
						romf.battery = true
					case 2:
						romf.trainer = true
					case 3:
						romf.fourScreen = true
						romf.mirroring = rp2cgo2.FourScreen
					}
				}
			}

			romf.mapper = (byte >> 4) & 0x0f
		case 7:
			if byte&0x01 != 0 {
				romf.vsCart = true
			}

			romf.mapper |= byte & 0xf0

		case 8:
			romf.ramBanks = byte

			if romf.ramBanks == 0 {
				romf.ramBanks = 1
			}
		case 9:
			if byte&0x01 != 0 {
				romf.region = PAL
			}
		}
	}

	i += 6

	if romf.trainer {
		offset = 512

		if len(buf) < (i + offset) {
			romf = nil
			err = errors.New("Invalid ROM: EOF in trainer data")
			return
		}

		romf.trainerData = buf[i : i+offset]
		i += offset
	}

	offset = 1024 * 16

	if len(buf) < (i + (offset * int(romf.prgBanks))) {
		romf = nil
		err = errors.New("Invalid ROM: EOF in ROM bank data")
		return
	}

	romf.romBanks = make([][]uint8, romf.prgBanks)

	for n := 0; n < int(romf.prgBanks); n++ {
		romf.romBanks[n] = buf[i : i+offset]
		i += offset
	}

	offset = 1024 * 8

	if len(buf) < (i + (offset * int(romf.chrBanks))) {
		romf = nil
		err = errors.New("Invalid ROM: EOF in VROM bank data")
		return
	}

	romf.vromBanks = make([][]uint8, romf.chrBanks)

	for n := 0; n < int(romf.chrBanks); n++ {
		romf.vromBanks[n] = buf[i : i+offset]
		i += offset
	}

	offset = 1024 * 8

	romf.wramBanks = make([][]uint8, romf.ramBanks)

	for n := 0; n < int(romf.ramBanks); n++ {
		romf.wramBanks[n] = make([]uint8, offset)
	}

	return
}

func (romf *ROMFile) Region() Region {
	return romf.region
}

func (romf *ROMFile) Tables() (t0, t1, t2, t3 int) {
	switch romf.mirroring {
	case rp2cgo2.Horizontal:
		t0, t1, t2, t3 = 0, 0, 1, 1
	case rp2cgo2.Vertical:
		t0, t1, t2, t3 = 0, 1, 0, 1
	}

	return
}

func (romf *ROMFile) String() string {
	return fmt.Sprintf("PRG Banks: %v\n", romf.prgBanks) +
		fmt.Sprintf("CHR Banks: %v\n", romf.chrBanks) +
		fmt.Sprintf("Mirroring: %v\n", romf.mirroring) +
		fmt.Sprintf("Battery: %v\n", romf.battery) +
		fmt.Sprintf("Trainer: %v\n", romf.trainer) +
		fmt.Sprintf("FourScreen: %v\n", romf.fourScreen) +
		fmt.Sprintf("VS Cart: %v\n", romf.vsCart) +
		fmt.Sprintf("RAM Banks: %v\n", romf.ramBanks) +
		fmt.Sprintf("Region: %v\n", romf.region)
}

func (romf *ROMFile) GameName() string {
	return romf.gamename
}

func (romf *ROMFile) LoadBattery() {
	var ram []byte

	if !romf.battery || romf.ramBanks == 0 {
		return
	}

	savename := romf.gamename + ".sav"
	ram, err := ioutil.ReadFile(savename)

	if err != nil {
		return
	}

	fmt.Println("*** Loading battery from " + savename)

	for b := range romf.wramBanks {
		for i := uint16(0); i < 0x2000; i++ {
			romf.wramBanks[b][i] = ram[i]
		}
	}

	return
}

func (romf *ROMFile) SaveBattery() (err error) {
	if !romf.battery || romf.ramBanks == 0 {
		return
	}

	savename := romf.gamename + ".sav"

	fmt.Println("*** Saving battery to " + savename)

	buf := bytes.Buffer{}

	for b := range romf.wramBanks {
		for i := uint16(0); i < 0x2000; i++ {
			buf.WriteByte(romf.wramBanks[b][i])
		}
	}

	err = ioutil.WriteFile(savename, buf.Bytes(), 0644)

	return
}
