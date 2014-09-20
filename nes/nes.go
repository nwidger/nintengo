package nes

import (
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
	state         RunState
	paused        chan bool
	events        chan Event
	cpu           *rp2ago3.RP2A03
	cpuDivisor    float32
	ppu           *rp2cgo2.RP2C02
	controllers   *Controllers
	rom           ROM
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
	video, err = NewSDLVideo(events)

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
		paused:        make(chan bool),
		events:        events,
		cpu:           cpu,
		cpuDivisor:    cpuDivisor,
		ppu:           ppu,
		rom:           rom,
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
	nes.cpu.Reset()
	nes.ppu.Reset()
	nes.controllers.Reset()
}

func (nes *NES) processEvents() {
	for nes.state != Quitting {
		e := <-nes.events
		e.Process(nes)
	}
}

func (nes *NES) runProcessors() (err error) {
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

			for quota += float32(cycles) * nes.cpuDivisor; quota >= 1.0; quota-- {
				if colors := nes.ppu.Execute(); colors != nil {
					nes.frame(colors)
					nes.fps.Delay()
				}
			}

			for i := uint16(0); i < cycles; i++ {
				if sample, haveSample := nes.cpu.APU.Execute(); haveSample {
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
	fmt.Println(nes.rom)

	nes.rom.LoadBattery()
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

	err = nes.rom.SaveBattery()

	return
}
