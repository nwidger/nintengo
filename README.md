nintengo
========

[![Build Status](https://travis-ci.org/nwidger/nintengo.svg?branch=master)](https://travis-ci.org/nwidger/nintengo)

An NES emulator written in Go

![Super Mario Bros.](http://i.imgur.com/g6ogqv7.gif "Super Mario Bros.")

![Donkey Kong](http://i.imgur.com/0SIbydD.gif "Donkey Kong")

![Excitebike](http://i.imgur.com/NTYlltB.gif "Excitebike")

![Legend of Zelda](http://i.imgur.com/XnrqFhI.gif "Legend of Zelda")

![Punch-Out!](http://i.imgur.com/UbIroEM.gif "Punch-Out!")

![Super Mario Bros. 3](http://i.imgur.com/bdXDNiY.gif "Super Mario Bros. 3")

![Mega Man 2](http://i.imgur.com/nZTU4i4.gif "Mega Man 2")

## Build

### Mac OS X/Linux/Windows

1. Install Azul3D by following the official
   [installation instructions](http://azul3d.org/doc/install) for your
   platform.

2. `go get -u github.com/nwidger/nintengo`

### WebAssembly

WebAssembly support requires Go 1.11 or higher.

1. Download nintengo

   ```
   go get -d -u github.com/nwidger/nintengo
   ```

2. Build `wasm/nintengo.wasm`

   ```
   cd $GOPATH/src/github.com/nwidger/nintengo
   GOOS=js GOARCH=wasm go build -o wasm/nintengo.wasm .
   ```

3. Start web server running on port 8000

   ```
   cd $GOPATH/src/github.com/nwidger/nintengo/wasm
   go run serve.go -http :8000
   ```

3. Open `http://localhost:8000` in your browser.  Press the `Choose
   File` button and select a `.nes` file to run.

## Usage

```
nintengo OPTIONS FILE
FILE can be a .nes file or a .nes file inside a .zip archive
  -audio-recorder="": recorder to use: none | wav
  -connect="": Connect to address as slave, <rom-file> will be ignored (e.g., 'localhost:8080')
  -cpu-decode=false: decode CPU instructions
  -cpu-profile="": write CPU profile to file
  -http="": HTTP service address (e.g., ':6060')
  -listen="": Listen at address as master (e.g., ':8080')
  -mem-profile="": write memory profile to file
  -recorder="": recorder to use: none | jpeg | gif
  -region="NTSC": system region to emulate: NTSC | PAL
```

## Controls

```
z - A
x - B
Enter - Start
Right Shift - Select
Arrow keys - Up/Down/Left/Right

p - Pause/Unpause
n - Toggle stepping by cycle/scanline/frame with p
r - Reset
q - Quit

F1 - save state
F5 - load state

F8  - 200% FPS (2x fast forward)
F9  - 100% FPS
F10 - 75% FPS
F11 - 50% FPS
F12 - 25% FPS

` - toggle overscan
1 - 256x240 screen size
2 - 512x480 screen size
3 - 768x720 screen size
4 - 1024x960 screen size
5 - 2560x1440 screen size

9 - Show/hide background
0 - Show/hide sprites

keypad 0 - toggle mute all channels
keypad 1 - toggle mute pulse 1 channel
keypad 2 - toggle mute pulse 2 channel
keypad 3 - toggle mute triangle channel
keypad 4 - toggle mute noise channel
keypad 5 - toggle mute DMC channel

l - Save pattern tables to left/right.jpg

o - Toggle CPU decoding
i - Toggle PPU decoding

with -recorder=gif:
s - Start recording to frame.gif
d - Stop recording

with -recorder=jpeg:
s - Save screenshot to frame.jpg

with -audio-recorder=wav:
keypad - (minus) - Start audio recording to audio.wav
keypad + (plus) - Stop audio recording
```

## Support

Audio support is currently a work in progress.  All audio channels are
working in some capacity.

Battery backed saves is implemented and are saved to disk with a
`.sav` file extension.

Save states are supported and are saved to disk with a `.nst` file
extension.

## Netplay

Nintengo includes two-player netplay support using the `-listen` and
`-connect` command-line arguments.  To use, player one launches
nintengo and tells it to listen for incoming connections using the
`-listen` flag:

```
nintengo -listen=:8080 FILE
```

Player two then connects to player one using the `-connect` flag,
providing the server's host/port:

```
nintengo -connect=192.168.1.110:8080
```

Note that player two does not need to provide a `FILE` argument.

## Mappers

- NROM
- MMC1
- UNROM
- CNROM
- MMC3
- ANROM
- MMC2

## Acknowledgments

This project would not have been possible without the amazing treasure
trove of information that is the
[NESdev Wiki](http://wiki.nesdev.com/w/index.php/NES_reference_guide)
and the fine folks at `#nesdev` on EFnet.

Being able to read the source code of other NES emulators was also
extremely helpful.  The following is a short list of the ones I
consulted most often, although it is probably not the complete list:

- Fergulator: https://github.com/scottferg/Fergulator/
- Nintendulator: http://www.qmtpro.com/~nes/nintendulator/
- Nestopia: http://nestopia.sourceforge.net/
- Nesalizer (by Ulf Magnusson): https://github.com/ulfalizer/nesalizer
