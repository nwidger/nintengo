nintengo
========

[![Build Status](https://travis-ci.org/nwidger/nintengo.svg?branch=master)](https://travis-ci.org/nwidger/nintengo)

An NES emulator written in Go

![Super Mario Bros.](http://i.imgur.com/g6ogqv7.gif "Super Mario Bros.")

![Donkey Kong](http://i.imgur.com/0SIbydD.gif "Donkey Kong")

![Excitebike](http://i.imgur.com/NTYlltB.gif "Excitebike")

![Legend of Zelda](http://i.imgur.com/XnrqFhI.gif "Legend of Zelda")

![Punch-Out!](http://i.imgur.com/UbIroEM.gif "Punch-Out!")

## Build

Requires Go 1.1.

- Linux

    ```
    $ sudo apt-get install libsdl1.2-dev libsdl-gfx1.2-dev libsdl-image1.2-dev libglew1.6-dev libxrandr-dev
    $ go get github.com/nwidger/nintengo
    ```

- Mac OS X

    ```
    $ brew install sdl --with-x11-driver
    $ brew install sdl_gfx sdl_image glew
    $ brew edit sdl
    $ go get github.com/nwidger/nintengo
    ```

## Usage

```
nintengo OPTIONS FILE
  -cpu-decode=false: decode CPU instructions
  -cpu-profile="": write CPU profile to file
  -mem-profile="": write memory profile to file
  -recorder="": recorder to use: none | jpeg | gif
```

## Controls

```
z - A
x - B
Enter - Start
Right Shift - Select
Arrow keys - Up/Down/Left/Right

p - Pause/Unpause
r - Reset
q - Quit

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

Audio support is currently a work in progress.  All audio channels
except the DMC channel are working in some capacity.

Battery backed saves is implemented.

### Mappers

- NROM
- MMC1
- UNROM
- CNROM
- MMC3
- ANROM
- MMC2
