package nes

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"runtime"

	"os"
	"runtime/pprof"

	"github.com/nwidger/nintengo/m65go2"
	"github.com/nwidger/nintengo/rp2ago3"
	"github.com/nwidger/nintengo/rp2cgo2"
)

type RunState uint8

const (
	Running RunState = 1 << iota
	Paused
	Quitting
)

type NES struct {
	state       RunState
	paused      chan bool
	events      chan Event
	cpu         *rp2ago3.RP2A03
	cpuDivisor  float32
	ppu         *rp2cgo2.RP2C02
	controllers *Controllers
	rom         ROM
	audio       Audio
	video       Video
	fps         *FPS
	recorder    Recorder
	options     *Options
}

type Options struct {
	Recorder   string
	CPUDecode  bool
	CPUProfile string
	MemProfile string
}

func NewNES(filename string, options *Options) (nes *NES, err error) {
	var audio Audio
	var video Video
	var recorder Recorder
	var cpuDivisor float32

	cpu := rp2ago3.NewRP2A03()

	if options.CPUDecode {
		cpu.EnableDecode()
	}

	rom, err := NewROM(filename, cpu.InterruptLine(m65go2.Irq))

	if err != nil {
		err = errors.New(fmt.Sprintf("Error loading ROM: %v", err))
		return
	}

	switch rom.Region() {
	case NTSC:
		cpuDivisor = rp2ago3.NTSC_CPU_CLOCK_DIVISOR
	case PAL:
		cpuDivisor = rp2ago3.PAL_CPU_CLOCK_DIVISOR
	}

	ctrls := NewControllers()

	events := make(chan Event)
	video, err = NewSDLVideo(events)

	if err != nil {
		err = errors.New(fmt.Sprintf("Error creating video: %v", err))
		return
	}

	audio, err = NewSDLAudio()

	if err != nil {
		err = errors.New(fmt.Sprintf("Error creating audio: %v", err))
		return
	}

	switch options.Recorder {
	case "none":
		// none
	case "jpeg":
		recorder, err = NewJPEGRecorder()
	case "gif":
		recorder, err = NewGIFRecorder()
	}

	if err != nil {
		err = errors.New(fmt.Sprintf("Error creating recorder: %v", err))
		return
	}

	ppu := rp2cgo2.NewRP2C02(cpu.InterruptLine(m65go2.Nmi))

	cpu.Memory.AddMappings(ppu, rp2ago3.CPU)
	cpu.Memory.AddMappings(rom, rp2ago3.CPU)
	cpu.Memory.AddMappings(ctrls, rp2ago3.CPU)

	ppu.Nametable.SetTables(rom.Tables())
	ppu.Memory.AddMappings(rom, rp2ago3.PPU)

	nes = &NES{
		paused:      make(chan bool),
		events:      events,
		cpu:         cpu,
		cpuDivisor:  cpuDivisor,
		ppu:         ppu,
		rom:         rom,
		audio:       audio,
		video:       video,
		fps:         NewFPS(DEFAULT_FPS),
		recorder:    recorder,
		controllers: ctrls,
		options:     options,
	}

	return
}

func (nes *NES) Reset() {
	nes.cpu.Reset()
	nes.ppu.Reset()
	nes.controllers.Reset()
}

func (nes *NES) SaveState() {
	fo, err := os.Create(fmt.Sprintf("game.save"))

	if err != nil {
		fmt.Println("*** Error saving state:", err)
	}

	defer fo.Close()

	w := bufio.NewWriter(fo)

	fmt.Println("*** Saving state")

	for i := uint32(0); i < 0x10000; i++ {
		err = w.WriteByte(nes.cpu.M6502.Memory.Fetch(uint16(i)))

		if err != nil {
			fmt.Println("*** Error saving state: CPU byte:", i, err)
		}
	}

	err = w.WriteByte(uint8(nes.cpu.M6502.Registers.PC >> 8))

	if err != nil {
		fmt.Println("*** Error saving state: High PC:", err)
	}

	err = w.WriteByte(uint8(nes.cpu.M6502.Registers.PC & 0xff))

	if err != nil {
		fmt.Println("*** Error saving state: Low PC:", err)
	}

	err = w.WriteByte(uint8(nes.cpu.M6502.Registers.A))

	if err != nil {
		fmt.Println("*** Error saving state: A:", err)
	}

	err = w.WriteByte(uint8(nes.cpu.M6502.Registers.X))

	if err != nil {
		fmt.Println("*** Error saving state: X:", err)
	}

	err = w.WriteByte(uint8(nes.cpu.M6502.Registers.Y))

	if err != nil {
		fmt.Println("*** Error saving state: Y:", err)
	}

	err = w.WriteByte(uint8(nes.cpu.M6502.Registers.P))

	if err != nil {
		fmt.Println("*** Error saving state: P:", err)
	}

	err = w.WriteByte(uint8(nes.cpu.M6502.Registers.SP))

	if err != nil {
		fmt.Println("*** Error saving state: SP:", err)
	}

	for i := uint32(0); i < 0x10000; i++ {
		err = w.WriteByte(nes.ppu.Memory.Memory.Fetch(uint16(i)))

		if err != nil {
			fmt.Println("*** Error saving state: PPU byte:", i, err)
		}
	}

	w.Flush()
}

func (nes *NES) LoadState() {
	fo, err := os.Open(fmt.Sprintf("game.save"))

	if err != nil {
		fmt.Println("*** Error loading state: Opening game.save:", err)
	}

	defer fo.Close()

	r := bufio.NewReader(fo)

	fmt.Println("*** Loading state")

	for i := uint32(0); i < 0x10000; i++ {
		b, err := r.ReadByte()

		if err != nil {
			fmt.Println("*** Error loading state: CPU byte:", i, err)
		}

		nes.cpu.M6502.Memory.Store(uint16(i), b)
	}

	high, err := r.ReadByte()

	if err != nil {
		fmt.Println("*** Error loading state: High PC:", err)
	}

	low, err := r.ReadByte()

	if err != nil {
		fmt.Println("*** Error loading state: Low PC:", err)
	}

	nes.cpu.M6502.Registers.PC = (uint16(high) << 8) | uint16(low)

	nes.cpu.M6502.Registers.A, err = r.ReadByte()

	if err != nil {
		fmt.Println("*** Error loading state: A:", err)
	}

	nes.cpu.M6502.Registers.X, err = r.ReadByte()

	if err != nil {
		fmt.Println("*** Error loading state: X:", err)
	}

	nes.cpu.M6502.Registers.Y, err = r.ReadByte()

	if err != nil {
		fmt.Println("*** Error loading state: Y:", err)
	}

	b, err := r.ReadByte()

	nes.cpu.M6502.Registers.P = m65go2.Status(b)

	if err != nil {
		fmt.Println("*** Error loading state: P:", err)
	}

	nes.cpu.M6502.Registers.SP, err = r.ReadByte()

	if err != nil {
		fmt.Println("*** Error loading state: SP:", err)
	}

	for i := uint32(0); i < 0x10000; i++ {
		b, err := r.ReadByte()

		if err != nil {
			fmt.Printf("*** Error loading state: PPU byte: %02x %v\n", i, err)
		}

		nes.cpu.Memory.Memory.Store(uint16(i), b)
	}
}

func (nes *NES) processEvents() {
	for nes.state != Quitting {
		e := <-nes.events
		go e.Process(nes)
	}
}

func (nes *NES) RunProcessors() (err error) {
	var cycles uint16

	quota := float32(0)

	for nes.state != Quitting {
		select {
		case paused := <-nes.paused:
			if paused {
				<-nes.paused
			}
		default:
			if cycles, err = nes.cpu.Execute(); err != nil {
				break
			}

			if nes.rom.RefreshTables() {
				nes.ppu.Nametable.SetTables(nes.rom.Tables())
			}

			for quota += float32(cycles) * nes.cpuDivisor; quota >= 1.0; quota-- {
				if colors := nes.ppu.Execute(); colors != nil {
					go nes.frame(colors)
					nes.fps.Delay()
				}
			}
		}
	}

	return
}

func (nes *NES) frame(colors []uint8) {
	nes.events <- &FrameEvent{
		colors: colors,
	}
}

func (nes *NES) Run() (err error) {
	fmt.Println(nes.rom)

	nes.rom.LoadBattery()
	nes.Reset()

	nes.state = Running

	go nes.RunProcessors()
	go nes.processEvents()

	if nes.recorder != nil {
		go nes.recorder.Run()
	}

	runtime.LockOSThread()
	runtime.GOMAXPROCS(runtime.NumCPU())

	if nes.options.CPUProfile != "" {
		f, err := os.Create(nes.options.CPUProfile)

		if err != nil {
			log.Fatal(err)
		}

		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	nes.video.Run()

	if nes.recorder != nil {
		nes.recorder.Quit()
	}

	if nes.options.MemProfile != "" {
		f, err := os.Create(nes.options.MemProfile)

		if err != nil {
			log.Fatal(err)
		}

		pprof.WriteHeapProfile(f)
		f.Close()
	}

	err = nes.rom.SaveBattery()

	return
}
