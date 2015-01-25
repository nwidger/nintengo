package nes

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
)

type Packet struct {
	Tick uint64
	Ev   Event
}

type Event interface {
	Process(nes *NES)
	String() string
}

func init() {
	gob.Register(&FrameEvent{})
	gob.Register(&SampleEvent{})
	gob.Register(&ControllerEvent{})
	gob.Register(&PauseEvent{})
	gob.Register(&FrameStepEvent{})
	gob.Register(&ResetEvent{})
	gob.Register(&RecordEvent{})
	gob.Register(&StopEvent{})
	gob.Register(&AudioRecordEvent{})
	gob.Register(&AudioStopEvent{})
	gob.Register(&QuitEvent{})
	gob.Register(&ShowBackgroundEvent{})
	gob.Register(&ShowSpritesEvent{})
	gob.Register(&CPUDecodeEvent{})
	gob.Register(&PPUDecodeEvent{})
	gob.Register(&SaveStateEvent{})
	gob.Register(&LoadStateEvent{})
	gob.Register(&FPSEvent{})
	gob.Register(&SavePatternTablesEvent{})
	gob.Register(&MuteEvent{})
	gob.Register(&MuteNoiseEvent{})
	gob.Register(&MuteTriangleEvent{})
	gob.Register(&MutePulse1Event{})
	gob.Register(&MutePulse2Event{})
	gob.Register(&HeartbeatEvent{})
}

const (
	EV_GLOBAL uint = 1 << iota
	EV_MASTER
	EV_SLAVE
)

func GetEventFlag(ev Event) (flag uint) {
	switch ev.String() {
	case "FrameEvent":
		flag |= EV_MASTER | EV_SLAVE
	case "SampleEvent":
		flag |= EV_MASTER | EV_SLAVE
	case "ControllerEvent":
		flag |= EV_GLOBAL | EV_MASTER | EV_SLAVE
	case "PauseEvent":
		flag |= EV_GLOBAL | EV_MASTER | EV_SLAVE
	case "FrameStepEvent":
		flag |= EV_GLOBAL | EV_MASTER
	case "ResetEvent":
		flag |= EV_GLOBAL | EV_MASTER
	case "RecordEvent":
		flag |= EV_MASTER | EV_SLAVE
	case "StopEvent":
		flag |= EV_MASTER | EV_SLAVE
	case "AudioRecordEvent":
		flag |= EV_MASTER | EV_SLAVE
	case "AudioStopEvent":
		flag |= EV_MASTER | EV_SLAVE
	case "QuitEvent":
		flag |= EV_MASTER | EV_SLAVE
	case "ShowBackgroundEvent":
		flag |= EV_MASTER | EV_SLAVE
	case "ShowSpritesEvent":
		flag |= EV_MASTER | EV_SLAVE
	case "CPUDecodeEvent":
		flag |= EV_MASTER | EV_SLAVE
	case "PPUDecodeEvent":
		flag |= EV_MASTER | EV_SLAVE
	case "SaveStateEvent":
		flag |= EV_GLOBAL | EV_MASTER
	case "LoadStateEvent":
		flag |= EV_GLOBAL | EV_MASTER
	case "FPSEvent":
		flag |= EV_GLOBAL | EV_MASTER
	case "SavePatternTablesEvent":
		flag |= EV_MASTER | EV_SLAVE
	case "MuteEvent":
		flag |= EV_MASTER | EV_SLAVE
	case "MuteNoiseEvent":
		flag |= EV_MASTER | EV_SLAVE
	case "MuteTriangleEvent":
		flag |= EV_MASTER | EV_SLAVE
	case "MutePulse1Event":
		flag |= EV_MASTER | EV_SLAVE
	case "MutePulse2Event":
		flag |= EV_MASTER | EV_SLAVE
	case "HeartbeatEvent":
		flag |= EV_GLOBAL | EV_MASTER
	default:
	}
	return
}

type FrameEvent struct {
	Colors []uint8
}

func (e *FrameEvent) String() string {
	return "FrameEvent"
}

func (e *FrameEvent) Process(nes *NES) {
	if nes.state != Running {
		return
	}

	if nes.recorder != nil {
		nes.recorder.Input() <- e.Colors
	}

	nes.video.Input() <- e.Colors
}

type SampleEvent struct {
	Sample int16
}

func (e *SampleEvent) String() string {
	return "SampleEvent"
}

func (e *SampleEvent) Process(nes *NES) {
	if nes.state != Running {
		return
	}

	if nes.audioRecorder != nil {
		nes.audioRecorder.Input() <- e.Sample
	}

	nes.audio.Input() <- e.Sample
}

type ControllerEvent struct {
	Controller int
	Down       bool
	B          Button
}

func (e *ControllerEvent) String() string {
	return "ControllerEvent"
}

func (e *ControllerEvent) Process(nes *NES) {
	if nes.state != Running {
		return
	}

	if e.Down {
		nes.controllers.KeyDown(e.Controller, e.B)
	} else {
		nes.controllers.KeyUp(e.Controller, e.B)
	}
}

type PauseEvent struct {
	Request PauseRequest
	Changed chan bool
}

func (e *PauseEvent) String() string {
	return "PauseEvent"
}

func (e *PauseEvent) Process(nes *NES) {
	nes.audio.TogglePaused()
	nes.paused <- e
}

type FrameStepEvent struct{}

func (e *FrameStepEvent) String() string {
	return "FrameStepEvent"
}

func (e *FrameStepEvent) Process(nes *NES) {
	switch nes.frameStep {
	case NoStep:
		fmt.Println("*** Press pause to step by cycle")
		nes.frameStep = CycleStep
	case CycleStep:
		fmt.Println("*** Press pause to step by scanline")
		nes.frameStep = ScanlineStep
	case ScanlineStep:
		fmt.Println("*** Press pause to step by frame")
		nes.frameStep = FrameStep
	case FrameStep:
		fmt.Println("*** Stepping disabled, press pause to continue")
		nes.frameStep = NoStep
	}
}

type ResetEvent struct{}

func (e *ResetEvent) String() string {
	return "ResetEvent"
}

func (e *ResetEvent) Process(nes *NES) {
	if nes.state != Running {
		return
	}

	nes.Reset()
}

type RecordEvent struct{}

func (e *RecordEvent) String() string {
	return "RecordEvent"
}

func (e *RecordEvent) Process(nes *NES) {
	if nes.recorder != nil {
		nes.recorder.Record()
	}
}

type StopEvent struct{}

func (e *StopEvent) String() string {
	return "StopEvent"
}

func (e *StopEvent) Process(nes *NES) {
	if nes.recorder != nil {
		nes.recorder.Stop()
	}
}

type AudioRecordEvent struct{}

func (e *AudioRecordEvent) String() string {
	return "AudioRecordEvent"
}

func (e *AudioRecordEvent) Process(nes *NES) {
	if nes.audioRecorder != nil {
		nes.audioRecorder.Record()
	}
}

type AudioStopEvent struct{}

func (e *AudioStopEvent) String() string {
	return "AudioStopEvent"
}

func (e *AudioStopEvent) Process(nes *NES) {
	if nes.audioRecorder != nil {
		nes.audioRecorder.Stop()
	}
}

type QuitEvent struct{}

func (e *QuitEvent) String() string {
	return "QuitEvent"
}

func (e *QuitEvent) Process(nes *NES) {
	nes.state = Quitting
}

type ShowBackgroundEvent struct{}

func (e *ShowBackgroundEvent) String() string {
	return "ShowBackgroundEvent"
}

func (e *ShowBackgroundEvent) Process(nes *NES) {
	nes.PPU.ShowBackground = !nes.PPU.ShowBackground
	fmt.Println("*** Toggling show background =", nes.PPU.ShowBackground)
}

type ShowSpritesEvent struct{}

func (e *ShowSpritesEvent) String() string {
	return "ShowSpritesEvent"
}

func (e *ShowSpritesEvent) Process(nes *NES) {
	nes.PPU.ShowSprites = !nes.PPU.ShowSprites
	fmt.Println("*** Toggling show sprites =", nes.PPU.ShowSprites)
}

type CPUDecodeEvent struct{}

func (e *CPUDecodeEvent) String() string {
	return "CPUDecodeEvent"
}

func (e *CPUDecodeEvent) Process(nes *NES) {
	fmt.Println("*** Toggling CPU decode =", nes.CPU.ToggleDecode())
}

type PPUDecodeEvent struct{}

func (e *PPUDecodeEvent) String() string {
	return "PPUDecodeEvent"
}

func (e *PPUDecodeEvent) Process(nes *NES) {
	fmt.Println("*** Toggling PPU decode =", nes.PPU.ToggleDecode())
}

type SaveStateEvent struct{}

func (e *SaveStateEvent) String() string {
	return "SaveStateEvent"
}

func (e *SaveStateEvent) Process(nes *NES) {
	nes.SaveState()
}

type LoadStateEvent struct {
	Data []byte
}

func (e *LoadStateEvent) String() string {
	return "LoadStateEvent"
}

func (e *LoadStateEvent) Process(nes *NES) {
	if e.Data == nil {
		if !nes.master {
			// Should not go here, NES.processEvents already filter the events.
			return
		}
		name := nes.ROM.GameName() + ".nst"
		fmt.Println("Reading ", name)
		data, err := ioutil.ReadFile(name)
		if err != nil {
			return
		}
		e.Data = data
	}
	reader := bytes.NewReader(e.Data)
	nes.LoadStateFromReader(reader, int64(len(e.Data)))
}

type FPSEvent struct {
	Rate float64
}

func (e *FPSEvent) String() string {
	return "FPSEvent"
}

func (e *FPSEvent) Process(nes *NES) {
	nes.fps.SetRate(DEFAULT_FPS * e.Rate)
	nes.audio.SetSpeed(1.00)
	fmt.Printf("*** Setting fps to %0.1f", e.Rate)
}

type SavePatternTablesEvent struct{}

func (e *SavePatternTablesEvent) String() string {
	return "SavePatternTablesEvent"
}

func (e *SavePatternTablesEvent) Process(nes *NES) {
	fmt.Println("*** Saving PPU pattern tables")
	nes.PPU.SavePatternTables()
}

type MuteEvent struct{}

func (e *MuteEvent) String() string {
	return "MuteEvent"
}

func (e *MuteEvent) Process(nes *NES) {
	nes.CPU.APU.Muted = !nes.CPU.APU.Muted
	fmt.Println("*** Toggling mute =", nes.CPU.APU.Muted)
}

type MuteNoiseEvent struct{}

func (e *MuteNoiseEvent) String() string {
	return "MuteNoiseEvent"
}

func (e *MuteNoiseEvent) Process(nes *NES) {
	nes.CPU.APU.Noise.Muted = !nes.CPU.APU.Noise.Muted
	fmt.Println("*** Toggling mute noise =", nes.CPU.APU.Noise.Muted)
}

type MuteTriangleEvent struct{}

func (e *MuteTriangleEvent) String() string {
	return "MuteTriangleEvent"
}

func (e *MuteTriangleEvent) Process(nes *NES) {
	nes.CPU.APU.Triangle.Muted = !nes.CPU.APU.Triangle.Muted
	fmt.Println("*** Toggling mute triangle =", nes.CPU.APU.Triangle.Muted)
}

type MutePulse1Event struct{}

func (e *MutePulse1Event) String() string {
	return "MutePulse1Event"
}

func (e *MutePulse1Event) Process(nes *NES) {
	nes.CPU.APU.Pulse1.Muted = !nes.CPU.APU.Pulse1.Muted
	fmt.Println("*** Toggling mute pulse1 =", nes.CPU.APU.Pulse1.Muted)
}

type MutePulse2Event struct{}

func (e *MutePulse2Event) String() string {
	return "MutePulse2Event"
}

func (e *MutePulse2Event) Process(nes *NES) {
	nes.CPU.APU.Pulse2.Muted = !nes.CPU.APU.Pulse2.Muted
	fmt.Println("*** Toggling mute pulse2 =", nes.CPU.APU.Pulse2.Muted)
}

type HeartbeatEvent struct{}

func (e *HeartbeatEvent) String() string {
	return "HeartbeatEvent"
}

func (e *HeartbeatEvent) Process(nes *NES) {
	// do nothing
}
