package nes

import "fmt"

type Event interface {
	Process(nes *NES)
}

type FrameEvent struct {
	colors []uint8
}

func (e *FrameEvent) String() string {
	return "FrameEvent"
}

func (e *FrameEvent) Process(nes *NES) {
	if nes.state != Running {
		return
	}

	if nes.recorder != nil {
		nes.recorder.Input() <- e.colors
	}

	nes.video.Input() <- e.colors
}

type SampleEvent struct {
	sample int16
}

func (e *SampleEvent) String() string {
	return "SampleEvent"
}

func (e *SampleEvent) Process(nes *NES) {
	if nes.state != Running {
		return
	}

	if nes.audioRecorder != nil {
		nes.audioRecorder.Input() <- e.sample
	}

	nes.audio.Input() <- e.sample
}

type ControllerEvent struct {
	controller int
	down       bool
	button     Button
}

func (e *ControllerEvent) String() string {
	return "ControllerEvent"
}

func (e *ControllerEvent) Process(nes *NES) {
	if nes.state != Running {
		return
	}

	if e.down {
		nes.controllers.KeyDown(e.controller, e.button)
	} else {
		nes.controllers.KeyUp(e.controller, e.button)
	}
}

type PauseEvent struct{}

func (e *PauseEvent) String() string {
	return "PauseEvent"
}

func (e *PauseEvent) Process(nes *NES) {
	nes.audio.TogglePaused()
	nes.paused <- true // send pause request
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
	pe := &PauseEvent{}
	state := nes.state

	if state == Running {
		pe.Process(nes)
	}

	nes.SaveState()

	if state == Running {
		pe.Process(nes)
	}
}

type LoadStateEvent struct{}

func (e *LoadStateEvent) String() string {
	return "LoadStateEvent"
}

func (e *LoadStateEvent) Process(nes *NES) {
	pe := &PauseEvent{}
	state := nes.state

	if state == Running {
		pe.Process(nes)
	}

	nes.LoadState()

	if state == Running {
		pe.Process(nes)
	}
}

type FastForwardEvent struct{}

func (e *FastForwardEvent) String() string {
	return "FastForwardEvent"
}

func (e *FastForwardEvent) Process(nes *NES) {
	nes.fps.SetRate(DEFAULT_FPS * 2.00)
	fmt.Println("*** Setting fps to fast forward (2x)")
}

type FPS100Event struct{}

func (e *FPS100Event) String() string {
	return "FPS100Event"
}

func (e *FPS100Event) Process(nes *NES) {
	nes.fps.SetRate(DEFAULT_FPS * 1.00)
	fmt.Println("*** Setting fps to 4/4")
}

type FPS75Event struct{}

func (e *FPS75Event) String() string {
	return "FPS75Event"
}

func (e *FPS75Event) Process(nes *NES) {
	nes.fps.SetRate(DEFAULT_FPS * 0.75)
	fmt.Println("*** Setting fps to 3/4")
}

type FPS50Event struct{}

func (e *FPS50Event) String() string {
	return "FPS50Event"
}

func (e *FPS50Event) Process(nes *NES) {
	nes.fps.SetRate(DEFAULT_FPS * 0.50)
	fmt.Println("*** Setting fps to 2/4")
}

type FPS25Event struct{}

func (e *FPS25Event) String() string {
	return "FPS25Event"
}

func (e *FPS25Event) Process(nes *NES) {
	nes.fps.SetRate(DEFAULT_FPS * 0.25)
	fmt.Println("*** Setting fps to 1/4")
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
