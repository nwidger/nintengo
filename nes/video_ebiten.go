package nes

import (
	"errors"
	"fmt"
	"image/color"
	"sync"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/inpututil"
)

type EbitenVideo struct {
	caption   string
	events    chan Event
	framePool *sync.Pool
	fps       float64
	input     chan []uint8
	overscan  bool
	frameMu   *sync.Mutex
	frame     *ebiten.Image
	palette   []color.Color
	stop      chan struct{}
}

func NewVideo(caption string, events chan Event, framePool *sync.Pool, fps float64) (video *EbitenVideo, err error) {
	video = &EbitenVideo{
		caption:   caption,
		events:    events,
		framePool: framePool,
		fps:       fps,
		input:     make(chan []uint8, 128),
		overscan:  true,
		frameMu:   &sync.Mutex{},
		palette:   RGBAPalette,
		stop:      make(chan struct{}),
	}
	err = video.setOverscan(false)
	if err != nil {
		return nil, err
	}
	return
}

func (video *EbitenVideo) Input() chan []uint8 {
	return video.input
}

func (video *EbitenVideo) Events() chan Event {
	return video.events
}

func (video *EbitenVideo) SetCaption(caption string) {
	video.caption = caption
}

func (video *EbitenVideo) Run() {
	defer close(video.stop)
	ebiten.SetMaxTPS(int(video.fps))
	ebiten.SetWindowSize(512, 480)
	ebiten.SetWindowResizable(true)
	ebiten.SetRunnableOnUnfocused(true)
	ebiten.SetCursorMode(ebiten.CursorModeHidden)
	go video.run()
	ebiten.RunGame(video)
}

func (video *EbitenVideo) run() {
	for {
		select {
		case colors := <-video.input:
			x, y := 0, 0

			video.frameMu.Lock()
			for _, c := range colors {
				if pixelInFrame(x, y, video.overscan) {
					video.frame.Set(x, y, video.palette[c])
				}

				switch x {
				case 255:
					x = 0
					y++
				default:
					x++
				}
			}
			video.framePool.Put(colors)
			video.frameMu.Unlock()
		case <-video.stop:
			break
		}
	}
}

func (video *EbitenVideo) setOverscan(overscan bool) error {
	video.frameMu.Lock()
	defer video.frameMu.Unlock()
	video.overscan = overscan
	frame, err := ebiten.NewImage(video.frameWidth(), video.frameHeight(), ebiten.FilterDefault)
	if err != nil {
		return err
	}
	video.frame = frame
	return nil
}

func (video *EbitenVideo) frameWidth() int {
	width := 256

	if video.overscan {
		width -= 16
	}

	return width
}

func (video *EbitenVideo) frameHeight() int {
	height := 240

	if video.overscan {
		height -= 16
	}

	return height
}

func (video *EbitenVideo) Update(_ *ebiten.Image) error {
	var event Event

	ebiten.SetWindowTitle(fmt.Sprintf("nintengo - %s - %.2f FPS", video.caption, ebiten.CurrentFPS()))

	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyGraveAccent):
		err := video.setOverscan(!video.overscan)
		if err != nil {
			return err
		}
	case inpututil.IsKeyJustPressed(ebiten.Key1):
		ebiten.SetWindowSize(256, 240)
	case inpututil.IsKeyJustPressed(ebiten.Key2):
		ebiten.SetWindowSize(512, 480)
	case inpututil.IsKeyJustPressed(ebiten.Key3):
		ebiten.SetWindowSize(768, 720)
	case inpututil.IsKeyJustPressed(ebiten.Key4):
		ebiten.SetWindowSize(1024, 960)
	case inpututil.IsKeyJustPressed(ebiten.Key5):
		ebiten.SetWindowSize(2560, 1440)
	case inpututil.IsKeyJustPressed(ebiten.KeyP):
		event = &PauseEvent{}
	case inpututil.IsKeyJustPressed(ebiten.KeyQ):
		event = &QuitEvent{}
		return errors.New("quit")
	case inpututil.IsKeyJustPressed(ebiten.KeyL):
		event = &SavePatternTablesEvent{}
	case inpututil.IsKeyJustPressed(ebiten.KeyR):
		event = &ResetEvent{}
	case inpututil.IsKeyJustPressed(ebiten.KeyS):
		event = &RecordEvent{}
	case inpututil.IsKeyJustPressed(ebiten.KeyD):
		event = &StopEvent{}
	case inpututil.IsKeyJustPressed(ebiten.KeyKPAdd):
		event = &AudioRecordEvent{}
	case inpututil.IsKeyJustPressed(ebiten.KeyKPSubtract):
		event = &AudioStopEvent{}
	case inpututil.IsKeyJustPressed(ebiten.KeyO):
		event = &CPUDecodeEvent{}
	case inpututil.IsKeyJustPressed(ebiten.KeyI):
		event = &PPUDecodeEvent{}
	case inpututil.IsKeyJustPressed(ebiten.Key9):
		event = &ShowBackgroundEvent{}
	case inpututil.IsKeyJustPressed(ebiten.Key0):
		event = &ShowSpritesEvent{}
	case inpututil.IsKeyJustPressed(ebiten.KeyF1):
		event = &SaveStateEvent{}
	case inpututil.IsKeyJustPressed(ebiten.KeyF5):
		event = &LoadStateEvent{}
	case inpututil.IsKeyJustPressed(ebiten.KeyF8):
		event = &FPSEvent{2.}
	case inpututil.IsKeyJustPressed(ebiten.KeyF9):
		event = &FPSEvent{1.}
	case inpututil.IsKeyJustPressed(ebiten.KeyF10):
		event = &FPSEvent{.75}
	case inpututil.IsKeyJustPressed(ebiten.KeyF11):
		event = &FPSEvent{.5}
	case inpututil.IsKeyJustPressed(ebiten.KeyF12):
		event = &FPSEvent{.25}
	case inpututil.IsKeyJustPressed(ebiten.KeyKP0):
		event = &MuteEvent{}
	case inpututil.IsKeyJustPressed(ebiten.KeyKP1):
		event = &MutePulse1Event{}
	case inpututil.IsKeyJustPressed(ebiten.KeyKP2):
		event = &MutePulse2Event{}
	case inpututil.IsKeyJustPressed(ebiten.KeyKP3):
		event = &MuteTriangleEvent{}
	case inpututil.IsKeyJustPressed(ebiten.KeyKP4):
		event = &MuteNoiseEvent{}
	case inpututil.IsKeyJustPressed(ebiten.KeyKP5):
		event = &MuteDMCEvent{}
	}

	if event == nil {
		button := One
		down := false

		switch {
		case inpututil.IsKeyJustPressed(ebiten.KeyZ) ||
			inpututil.IsKeyJustReleased(ebiten.KeyZ):
			button = A
			down = !inpututil.IsKeyJustReleased(ebiten.KeyZ)
		case inpututil.IsKeyJustPressed(ebiten.KeyX) ||
			inpututil.IsKeyJustReleased(ebiten.KeyX):
			button = B
			down = !inpututil.IsKeyJustReleased(ebiten.KeyX)
		case inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
			inpututil.IsKeyJustReleased(ebiten.KeyEnter):
			button = Start
			down = !inpututil.IsKeyJustReleased(ebiten.KeyEnter)
		case inpututil.IsKeyJustPressed(ebiten.KeyShift) ||
			inpututil.IsKeyJustReleased(ebiten.KeyShift):
			button = Select
			down = !inpututil.IsKeyJustReleased(ebiten.KeyShift)
		case inpututil.IsKeyJustPressed(ebiten.KeyUp) ||
			inpututil.IsKeyJustReleased(ebiten.KeyUp):
			button = Up
			down = !inpututil.IsKeyJustReleased(ebiten.KeyUp)
		case inpututil.IsKeyJustPressed(ebiten.KeyDown) ||
			inpututil.IsKeyJustReleased(ebiten.KeyDown):
			button = Down
			down = !inpututil.IsKeyJustReleased(ebiten.KeyDown)
		case inpututil.IsKeyJustPressed(ebiten.KeyLeft) ||
			inpututil.IsKeyJustReleased(ebiten.KeyLeft):
			button = Left
			down = !inpututil.IsKeyJustReleased(ebiten.KeyLeft)
		case inpututil.IsKeyJustPressed(ebiten.KeyRight) ||
			inpututil.IsKeyJustReleased(ebiten.KeyRight):
			button = Right
			down = !inpututil.IsKeyJustReleased(ebiten.KeyRight)
		}

		event = &ControllerEvent{
			Button: button,
			Down:   down,
		}
	}

	if event != nil {
		video.events <- event
	}

	return nil
}

func (video *EbitenVideo) Draw(screen *ebiten.Image) {
	video.frameMu.Lock()
	defer video.frameMu.Unlock()

	err := screen.DrawImage(video.frame, nil)
	if err != nil {
		panic("unable to draw image: " + err.Error())
	}
}

func (video *EbitenVideo) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 256, 240
}
