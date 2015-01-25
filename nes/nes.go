package nes

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"

	"os"
	"runtime"
	"runtime/pprof"

	"encoding/json"

	"archive/zip"

	"github.com/kaicheng/nintengo/m65go2"
	"github.com/kaicheng/nintengo/rp2ago3"
	"github.com/kaicheng/nintengo/rp2cgo2"
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
	state         RunState
	frameStep     StepState
	paused        chan *PauseEvent
	events        chan Event
	CPU           *rp2ago3.RP2A03
	cpuDivisor    float32
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

func NewNES(filename string, options *Options) (nes *NES, err error) {
	var audio Audio
	var video Video
	var recorder Recorder
	var audioRecorder AudioRecorder
	var cpuDivisor float32
	var master bool
	var bridge *Bridge

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
	}

	rom, err := NewROM(filename, cpu.InterruptLine(m65go2.Irq), ppu.Nametable.SetTables)

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
	video, err = NewVideo(rom.GameName(), events)

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
	cpu.Memory.AddMappings(rom, rp2ago3.CPU)
	cpu.Memory.AddMappings(ctrls, rp2ago3.CPU)

	ppu.Memory.AddMappings(rom, rp2ago3.PPU)

	lock := make(chan uint64, 1)
	lock <- 0

	nes = &NES{
		frameStep:     NoStep,
		paused:        make(chan *PauseEvent),
		events:        events,
		CPU:           cpu,
		cpuDivisor:    cpuDivisor,
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
	name := nes.ROM.GameName() + ".nst"

	fo, err := os.Create(name)
	defer fo.Close()

	if err != nil {
		fmt.Printf("*** Error saving state: %s\n", err)
		return
	}

	nes.SaveStateToWriter(fo)

	fmt.Println("*** Saving state to", name)
}

func (nes *NES) SaveStateToWriter(writer io.Writer) (err error) {
	fmt.Println("Start saving")
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

	zfw, err := zw.Create("state.json")

	if err != nil {
		fmt.Printf("*** Error saving state: %s\n", err)
		return
	}

	fmt.Println("Marshal nes")
	buf, err := json.MarshalIndent(nes, "", "  ")
	fmt.Println("Done marshal nes")

	if _, err = zfw.Write(buf); err != nil {
		fmt.Printf("*** Error saving state: %s\n", err)
	}
	return
}

func (nes *NES) LoadState() {
	name := nes.ROM.GameName() + ".nst"
	reader, err := os.Open(name)
	if err != nil {
	}
	defer reader.Close()
	readeri, err := reader.Stat()
	if err != nil {
	}

	nes.LoadStateFromReader(reader, readeri.Size())

	fmt.Println("*** Loading state from", name)
}

func (nes *NES) LoadStateFromReader(reader io.ReaderAt, size int64) (err error) {
	fmt.Println("Start loading state")
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

			loaded = true
		}
	}

	if !loaded {
		fmt.Printf("*** Error loading state: invalid save state file\n")
	}

	return
}

func (nes *NES) getLoadStateEvent() (ev *LoadStateEvent, err error) {
	var buf bytes.Buffer
	err = nes.SaveStateToWriter(&buf)
	if err != nil {
		fmt.Print("getLoadStateEvent: ", err)
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
		flag := GetEventFlag(e)
		if nes.master || flag&EV_SLAVE != 0 {
			if flag&EV_GLOBAL != 0 {
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
				fmt.Println("master? ", nes.master, ": Pkt into loop ", pkt)
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

		nes.PPUQuota += float32(cycles) * nes.cpuDivisor
	}

	if nes.PPUQuota >= 1.0 {
		scanline := nes.PPU.Scanline

		if colors := nes.PPU.Execute(); colors != nil {
			nes.frame(colors)
			nes.fps.Delay()

			if nes.frameStep == FrameStep {
				// isPaused = true
				fmt.Println("*** Paused at frame", nes.PPU.Frame)
			}
		}

		if mmc3 != nil && nes.PPU.TriggerScanlineCounter() {
			mmc3.scanlineCounter()
		}

		nes.PPUQuota--

		if nes.frameStep == CycleStep ||
			(nes.frameStep == ScanlineStep && nes.PPU.Scanline != scanline) {
			// isPaused = true

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
		fmt.Println("Got pkt: ", pkt)
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
				// error here.
			}
			nes.lock <- lock
		}
		pkt.Ev.Process(nes)
	}
	return
}

func (nes *NES) runAsMaster() (err error) {
	var cycles uint16

	// isPaused := false
	// mmc3, _ := nes.ROM.(*MMC3)

	nes.state = Running

	for nes.state != Quitting {
	ProcessingEventLoop:
		for {
			// Must be non-blocking receiving here
			select {
			case pkt := <-nes.bridge.incoming:
				fmt.Println("Sync processing: ", pkt)
				lock := <-nes.lock
				pkt.Tick = nes.Tick
				pkt.Ev.Process(nes)
				nes.lock <- lock
				flag := GetEventFlag(pkt.Ev)
				if nes.bridge.active && (flag&EV_GLOBAL != 0) {
					nes.bridge.outgoing <- pkt
				}
			default:
				// Done processing
				break ProcessingEventLoop
			}
		}

		lock := <-nes.lock
		cycles, err = nes.step()
		nes.Tick += uint64(cycles)

		nes.lock <- lock

		// FIXME: pausing
		/*
			select {
			case pr := <-nes.paused:
				isPaused = nes.isPaused(pr, isPaused)
			default:
			}

			for isPaused {
				isPaused = nes.isPaused(<-nes.paused, isPaused)
			}
		*/
	}
	return
}

func (nes *NES) isPaused(pr *PauseEvent, oldPaused bool) (isPaused bool) {
	switch pr.Request {
	case Pause:
		isPaused = true
	case Unpause:
		isPaused = false
	case Toggle:
		isPaused = !oldPaused
	}

	if pr.Changed != nil {
		pr.Changed <- (isPaused != oldPaused)
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

	nes.ROM.LoadBattery()
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

	err = nes.ROM.SaveBattery()

	return
}
