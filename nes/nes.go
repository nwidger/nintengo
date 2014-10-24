package nes

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"runtime"

	"os"
	"runtime/pprof"

	"encoding/json"

	"archive/zip"

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
	state         RunState
	paused        chan bool
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
}

type Options struct {
	Recorder      string
	AudioRecorder string
	CPUDecode     bool
	CPUProfile    string
	MemProfile    string
}

func NewNES(filename string, options *Options) (nes *NES, err error) {
	var audio Audio
	var video Video
	var recorder Recorder
	var audioRecorder AudioRecorder
	var cpuDivisor float32

	audioFrequency := 44100
	audioSampleSize := 2048

	cpu := rp2ago3.NewRP2A03(audioFrequency)

	if options.CPUDecode {
		cpu.EnableDecode()
	}

	ppu := rp2cgo2.NewRP2C02(cpu.InterruptLine(m65go2.Nmi))

	rom, err := NewROM(filename, cpu.InterruptLine(m65go2.Irq), ppu.SetTablesFunc())

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
	video, err = NewSDLVideo(rom.GameName(), events)

	if err != nil {
		err = errors.New(fmt.Sprintf("Error creating video: %v", err))
		return
	}

	audio, err = NewSDLAudio(audioFrequency, audioSampleSize)

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
	ppu.Memory.AddTracer(rom)

	nes = &NES{
		paused:        make(chan bool, 2),
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
	}

	return
}

func (nes *NES) Reset() {
	nes.CPU.Reset()
	nes.PPU.Reset()
	nes.PPUQuota = float32(0)
	nes.controllers.Reset()
}

func (nes *NES) SaveState() {
	name := nes.ROM.GameName() + ".nst"

	fo, err := os.Create(name)
	defer fo.Close()

	if err != nil {
		fmt.Printf("*** Error saving state: %s\n", err)
		return
	}

	w := bufio.NewWriter(fo)
	defer w.Flush()

	zw := zip.NewWriter(w)
	defer zw.Close()

	vfw, err := zw.Create("meta.json")

	if err != nil {
		fmt.Printf("*** Error saving state: %s\n", err)
		return
	}

	enc := json.NewEncoder(vfw)

	if err = enc.Encode(struct{ Version string }{"0.1"}); err != nil {
		fmt.Printf("*** Error saving state: %s\n", err)
		return
	}

	buf, err := json.MarshalIndent(nes, "", "  ")

	if err = enc.Encode(nes); err != nil {
		fmt.Printf("*** Error saving state: %s\n", err)
		return
	}

	zfw, err := zw.Create("state.json")

	if err != nil {
		fmt.Printf("*** Error saving state: %s\n", err)
		return
	}

	if _, err = zfw.Write(buf); err != nil {
		fmt.Printf("*** Error saving state: %s\n", err)
		return
	}

	fmt.Println("*** Saving state to", name)
}

func (nes *NES) LoadState() {
	name := nes.ROM.GameName() + ".nst"

	zr, err := zip.OpenReader(name)
	defer zr.Close()

	if err != nil {
		fmt.Printf("*** Error loading state: %s\n", err)
		return
	}

	loaded := false

	for _, zf := range zr.File {
		if zf.Name != "state.json" {
			continue
		}

		zfr, err := zf.Open()
		defer zfr.Close()

		if err != nil {
			fmt.Printf("*** Error loading state: %s\n", err)
			return
		}

		dec := json.NewDecoder(zfr)

		if err = dec.Decode(nes); err != nil {
			fmt.Printf("*** Error loading state: %s\n", err)
			return
		}

		loaded = true
	}

	if !loaded {
		fmt.Printf("*** Error loading state: invalid save state file\n")
		return
	}

	fmt.Println("*** Loading state from", name)
}

func (nes *NES) processEvents() {
	for nes.state != Quitting {
		e := <-nes.events
		e.Process(nes)
	}
}

func (nes *NES) runProcessors() (err error) {
	var cycles uint16

	for nes.state != Quitting {
		select {
		case paused := <-nes.paused:
			if paused {
				<-nes.paused
			}
		default:
			if cycles, err = nes.CPU.Execute(); err != nil {
				break
			}

			for nes.PPUQuota += float32(cycles) * nes.cpuDivisor; nes.PPUQuota >= 1.0; nes.PPUQuota-- {
				if colors := nes.PPU.Execute(); colors != nil {
					nes.frame(colors)
					nes.fps.Delay()
				}
			}

			for i := uint16(0); i < cycles; i++ {
				if sample, haveSample := nes.CPU.APU.Execute(); haveSample {
					nes.sample(sample)
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

func (nes *NES) sample(sample int16) {
	nes.events <- &SampleEvent{
		sample: sample,
	}
}

func (nes *NES) Run() (err error) {
	fmt.Println(nes.ROM)

	nes.ROM.LoadBattery()
	nes.Reset()

	nes.state = Running

	go nes.audio.Run()
	go nes.runProcessors()
	go nes.processEvents()

	if nes.recorder != nil {
		go nes.recorder.Run()
	}

	if nes.audioRecorder != nil {
		go nes.audioRecorder.Run()
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
