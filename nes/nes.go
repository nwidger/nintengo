package nes

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"

	"os"
	"runtime"
	"runtime/pprof"

	"encoding/json"

	"archive/zip"

	"github.com/nwidger/nintengo/m65go2"
	"github.com/nwidger/nintengo/rp2ago3"
	"github.com/nwidger/nintengo/rp2cgo2"
)

//go:generate stringer -type=StepState
type StepState uint8

const (
	NoStep StepState = iota
	CycleStep
	ScanlineStep
	FrameStep
)

//go:generate stringer -type=RunState
type RunState uint8

const (
	Uninitialized RunState = iota
	Running
	Quitting
)

type PauseRequest uint8

const (
	Toggle PauseRequest = iota
	Pause
	Unpause
)

type NES struct {
	GameName      string
	state         RunState
	frameStep     StepState
	Paused        bool
	events        chan Event
	CPU           *rp2ago3.RP2A03
	CpuDivisor    float32
	PPU           *rp2cgo2.RP2C02
	PPUQuota      float32
	controllers   *Controllers
	ROM           ROM
	audio         Audio
	video         Video
	fps           *FPS
	recorder      Recorder
	audioRecorder AudioRecorder
	options       *Options
	lock          chan uint64
	Tick          uint64
	master        bool
	bridge        *Bridge
}

type Options struct {
	Recorder      string
	AudioRecorder string
	CPUDecode     bool
	CPUProfile    string
	MemProfile    string
	HTTPAddress   string
	Listen        string
	Connect       string
}

func init() {
	gob.Register(&ROMFileRaw{})
}

func NewNES(filename string, options *Options) (nes *NES, err error) {
	var audio Audio
	var video Video
	var recorder Recorder
	var audioRecorder AudioRecorder
	var cpuDivisor float32
	var master bool
	var bridge *Bridge
	var rom ROM

	gamename := "NONAME"

	audioFrequency := 44100
	audioSampleSize := 2048

	cpu := rp2ago3.NewRP2A03(audioFrequency)

	if options.CPUDecode {
		cpu.EnableDecode()
	}

	ppu := rp2cgo2.NewRP2C02(cpu.InterruptLine(m65go2.Nmi))

	if len(options.Connect) > 0 {
		master = false
		bridge = newBridge(nil, options.Connect)
	} else {
		master = true
		bridge = newBridge(nil, options.Listen)

		rom, err = NewROM(filename, cpu.InterruptLine(m65go2.Irq), ppu.Nametable.SetTables)
		if err != nil {
			err = errors.New(fmt.Sprintf("Error loading ROM: %v", err))
			return
		}
		gamename = rom.GameName()
		switch rom.Region() {
		case NTSC:
			cpuDivisor = rp2ago3.NTSC_CPU_CLOCK_DIVISOR
		case PAL:
			cpuDivisor = rp2ago3.PAL_CPU_CLOCK_DIVISOR
		}

	}

	ctrls := NewControllers()

	events := make(chan Event)
	video, err = NewVideo(gamename, events)

	if err != nil {
		err = errors.New(fmt.Sprintf("Error creating video: %v", err))
		return
	}

	audio, err = NewAudio(audioFrequency, audioSampleSize)

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

	switch options.AudioRecorder {
	case "none":
		// none
	case "wav":
		audioRecorder, err = NewWAVRecorder()
	}

	if err != nil {
		err = errors.New(fmt.Sprintf("Error creating audio recorder: %v", err))
		return
	}

	cpu.Memory.AddMappings(ppu, rp2ago3.CPU)
	cpu.Memory.AddMappings(ctrls, rp2ago3.CPU)

	if master {
		cpu.Memory.AddMappings(rom, rp2ago3.CPU)
		ppu.Memory.AddMappings(rom, rp2ago3.PPU)
	}

	lock := make(chan uint64, 1)
	lock <- 0

	nes = &NES{
		GameName:      gamename,
		frameStep:     NoStep,
		events:        events,
		CPU:           cpu,
		CpuDivisor:    cpuDivisor,
		PPU:           ppu,
		ROM:           rom,
		audio:         audio,
		video:         video,
		fps:           NewFPS(DEFAULT_FPS),
		recorder:      recorder,
		audioRecorder: audioRecorder,
		controllers:   ctrls,
		options:       options,
		lock:          lock,
		Tick:          0,
		master:        master,
		bridge:        bridge,
	}

	bridge.nes = nes

	return
}

func (nes *NES) Reset() {
	nes.CPU.Reset()
	nes.PPU.Reset()
	nes.PPUQuota = float32(0)
	nes.controllers.Reset()
}

func (nes *NES) RunState() RunState {
	return nes.state
}

func (nes *NES) StepState() StepState {
	return nes.frameStep
}

func (nes *NES) Pause() RunState {
	e := &PauseEvent{}
	e.Process(nes)

	return nes.state
}

func (nes *NES) SaveState() {
	name := nes.GameName + ".nst"

	fo, err := os.Create(name)
	defer fo.Close()

	if err != nil {
		fmt.Printf("*** Error saving state: %s\n", err)
		return
	}

	fmt.Println("*** Saving state to", name)
	nes.SaveStateToWriter(fo, false)
}

func (nes *NES) SaveStateToWriter(writer io.Writer, withrom bool) (err error) {
	var romfw io.Writer

	w := bufio.NewWriter(writer)
	defer w.Flush()

	zw := zip.NewWriter(w)
	defer zw.Close()

	vfw, err := zw.Create("meta.json")

	if err != nil {
		fmt.Printf("*** Error saving state: %s\n", err)
		return
	}

	enc := json.NewEncoder(vfw)

	if err = enc.Encode(struct{ Version string }{"0.2"}); err != nil {
		fmt.Printf("*** Error saving state: %s\n", err)
		return
	}

	if withrom {
		// ROM should be written before NES, because we need ROM restored before NES.
		romfw, err = zw.Create("rom.bin")
		if err != nil {
			fmt.Printf("*** Error saving rom: %s\n", err)
			return
		}
		enc := gob.NewEncoder(romfw)
		raw := nes.ROM.GetRaw()
		if err = enc.Encode(raw); err != nil {
			fmt.Printf("*** Error saving rom: %s\n", err)
		}
	}

	zfw, err := zw.Create("state.json")

	if err != nil {
		fmt.Printf("*** Error saving state: %s\n", err)
		return
	}

	buf, err := json.MarshalIndent(nes, "", "  ")

	if _, err = zfw.Write(buf); err != nil {
		fmt.Printf("*** Error saving state: %s\n", err)
	}

	return
}

func (nes *NES) LoadState() {
	name := nes.GameName + ".nst"
	reader, err := os.Open(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file %s: %s\n", name, err)
	}
	defer reader.Close()
	readeri, err := reader.Stat()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting stat of %s: %s\n", name, err)
	}

	fmt.Println("*** Loading state from", name)
	nes.LoadStateFromReader(reader, readeri.Size())
}

func (nes *NES) LoadStateFromReader(reader io.ReaderAt, size int64) (err error) {
	zr, err := zip.NewReader(reader, size)

	if err != nil {
		fmt.Printf("*** Error loading state: %s\n", err)
		return
	}

	loaded := false

	for _, zf := range zr.File {
		switch zf.Name {
		case "meta.json":
			zfr, err := zf.Open()
			defer zfr.Close()

			if err != nil {
				fmt.Printf("*** Error loading state: %s\n", err)
				return err
			}

			dec := json.NewDecoder(zfr)

			v := struct{ Version string }{}

			if err = dec.Decode(&v); err != nil {
				fmt.Printf("*** Error loading state: %s\n", err)
				return err
			}

			if v.Version != "0.2" {
				fmt.Printf("*** Error loading state: Invalid save state format version '%s'\n", v.Version)
				return err
			}
		case "state.json":
			zfr, err := zf.Open()
			defer zfr.Close()

			if err != nil {
				fmt.Printf("*** Error loading state: %s\n", err)
				return err
			}

			dec := json.NewDecoder(zfr)

			if err = dec.Decode(nes); err != nil {
				fmt.Printf("*** Error loading state: %s\n", err)
				return err
			}
			nes.video.SetCaption(nes.GameName)

			loaded = true
		case "rom.bin":
			romfr, err := zf.Open()
			defer romfr.Close()

			if err != nil {
				fmt.Printf("*** Error loading rom: %s\n", err)
				return err
			}

			dec := gob.NewDecoder(romfr)
			var raw ROMFileRaw
			if err = dec.Decode(&raw); err != nil {
				fmt.Printf("*** Error loading rom %s\n", err)
				return err
			}

			rom, err := NewROMFromRaw(nes.GameName, raw.Raw, nes.CPU.InterruptLine(m65go2.Irq), nes.PPU.Nametable.SetTables)
			if err != nil {
				fmt.Printf("*** Error loading rom: %s\n", err)
				return err
			}

			romf := rom.GetROMFile()
			romf.wramBanks = raw.WRAMBanks
			romf.romBanks = raw.ROMBanks
			romf.vromBanks = raw.VROMBanks

			nes.ROM = rom
			nes.CPU.Memory.AddMappings(rom, rp2ago3.CPU)
			nes.PPU.Memory.AddMappings(rom, rp2ago3.PPU)
		}
	}

	if !loaded {
		fmt.Printf("*** Error loading state: invalid save state file\n")
	}

	return
}

func (nes *NES) getLoadStateEvent() (ev *LoadStateEvent, err error) {
	var buf bytes.Buffer
	err = nes.SaveStateToWriter(&buf, true)
	if err != nil {
		// Error message has been printed out.
		return
	}
	ev = &LoadStateEvent{
		Data: buf.Bytes(),
	}
	return
}

func (nes *NES) processEvents() {
	for nes.state != Quitting {
		e := <-nes.events
		flag := e.Flag()
		if nes.master || flag&EvSlave != 0 {
			if flag&EvGlobal != 0 {
				// Tick is not important here. Just a Reference
				if e.String() == "ControllerEvent" && !nes.master {
					// hardcode to fix controller id.
					ce, _ := e.(*ControllerEvent)
					ce.Controller = 1
				}
				pkt := Packet{
					Tick: nes.Tick,
					Ev:   e,
				}
				if nes.master {
					nes.bridge.incoming <- pkt
				} else {
					nes.bridge.outgoing <- pkt
				}
			} else {
				e.Process(nes)
			}
		}
	}
}

func (nes *NES) runProcessors() (err error) {
	if nes.master {
		return nes.runAsMaster()
	} else {
		return nes.runAsSlave()
	}
}

func (nes *NES) step() (cycles uint16, err error) {
	cycles = 0
	mmc3, _ := nes.ROM.(*MMC3)
	if nes.PPUQuota < 1.0 {
		if cycles, err = nes.CPU.Execute(); err != nil {
			return
		}

		nes.PPUQuota += float32(cycles) * nes.CpuDivisor
	}

	for nes.PPUQuota >= 1.0 {
		scanline := nes.PPU.Scanline

		if colors := nes.PPU.Execute(); colors != nil {
			nes.frame(colors)
			nes.fps.Delay()

			if nes.frameStep == FrameStep {
				nes.Paused = true
				fmt.Println("*** Paused at frame", nes.PPU.Frame)
			}
		}

		if mmc3 != nil && nes.PPU.TriggerScanlineCounter() {
			mmc3.scanlineCounter()
		}

		nes.PPUQuota--

		if nes.frameStep == CycleStep ||
			(nes.frameStep == ScanlineStep && nes.PPU.Scanline != scanline) {
			nes.Paused = true

			if nes.frameStep == CycleStep {
				fmt.Println("*** Paused at cycle", nes.PPU.Cycle)
			} else {
				fmt.Println("*** Paused at scanline", nes.PPU.Scanline)
			}
		}
	}

	if nes.PPUQuota < 1.0 {
		for i := uint16(0); i < cycles; i++ {
			if sample, haveSample := nes.CPU.APU.Execute(); haveSample {
				nes.sample(sample)
			}
		}
	}
	return
}

func (nes *NES) runAsSlave() (err error) {
	var cycles uint16
	for nes.state != Quitting {
		pkt := <-nes.bridge.incoming
		if pkt.Ev.String() != "LoadStateEvent" {
			if nes.state == Uninitialized {
				continue
			}
			lock := <-nes.lock
			for nes.Tick < pkt.Tick && err == nil {
				cycles, err = nes.step()
				nes.Tick += uint64(cycles)
			}
			if nes.Tick > pkt.Tick {
				fmt.Fprintf(os.Stderr, "Failed to sync with master, quiting...\n")
				err = errors.New(fmt.Sprintf("Failed to sync with master"))
				return
			}
			nes.lock <- lock
		}
		pkt.Ev.Process(nes)
	}
	return
}

func (nes *NES) processPacket(pkt *Packet) {
	lock := <-nes.lock
	pkt.Tick = nes.Tick
	pkt.Ev.Process(nes)
	nes.lock <- lock
	flag := pkt.Ev.Flag()
	if nes.bridge.active && (flag&EvGlobal != 0) {
		nes.bridge.outgoing <- *pkt
	}
}

func (nes *NES) runAsMaster() (err error) {
	var cycles uint16

	nes.state = Running

	for nes.state != Quitting {
	ProcessingEventLoop:
		for {
			if nes.Paused {
				// If Paused, use blocking chan receiving
				pkt := <-nes.bridge.incoming
				nes.processPacket(&pkt)
			} else {
				// Must be non-blocking receiving if not paused
				select {
				case pkt := <-nes.bridge.incoming:
					nes.processPacket(&pkt)
				default:
					// Done processing
					break ProcessingEventLoop
				}
			}
		}

		if !nes.Paused {
			lock := <-nes.lock
			cycles, err = nes.step()
			nes.Tick += uint64(cycles)

			nes.lock <- lock
		}
	}
	return
}

func (nes *NES) frame(colors []uint8) {

	// Generate a heartbeat each frame
	if nes.bridge.active {
		nes.bridge.outgoing <- Packet{
			Tick: nes.Tick,
			Ev:   &HeartbeatEvent{},
		}
	}

	e := &FrameEvent{
		Colors: colors,
	}

	e.Process(nes)
}

func (nes *NES) sample(sample int16) {
	e := &SampleEvent{
		Sample: sample,
	}

	e.Process(nes)
}

func (nes *NES) Run() (err error) {
	fmt.Println(nes.ROM)

	if nes.master {
		nes.ROM.LoadBattery()
	}
	nes.Reset()

	nes.state = Running

	go nes.audio.Run()
	go nes.processEvents()

	go func() {
		if err := nes.runProcessors(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}()

	if nes.recorder != nil {
		go nes.recorder.Run()
	}

	if nes.audioRecorder != nil {
		go nes.audioRecorder.Run()
	}

	if nes.master {
		go nes.bridge.runAsMaster()
	} else {
		go nes.bridge.runAsSlave()
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

	if nes.audioRecorder != nil {
		nes.audioRecorder.Quit()
	}

	if nes.options.MemProfile != "" {
		f, err := os.Create(nes.options.MemProfile)

		if err != nil {
			log.Fatal(err)
		}

		pprof.WriteHeapProfile(f)
		f.Close()
	}

	if nes.master {
		err = nes.ROM.SaveBattery()
	}

	return
}
