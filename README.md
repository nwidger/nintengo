nintengo
========

[![Build Status](https://travis-ci.org/nwidger/nintengo.svg?branch=master)](https://travis-ci.org/nwidger/nintengo)

An NES emulator written in Go

![Super Mario Bros.](http://i.imgur.com/g6ogqv7.gif "Super Mario Bros.")

![Donkey Kong](http://i.imgur.com/0SIbydD.gif "Donkey Kong")

![Excitebike](http://i.imgur.com/NTYlltB.gif "Excitebike")

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

F9  - 100% FPS
F10 - 75% FPS
F11 - 50% FPS
F12 - 25% FPS

1 - 1:1 aspect ratio
2 - 2:1 aspect ratio
3 - 3:1 aspect ratio
4 - 4:1 aspect ratio

9 - Show/hide background
0 - Show/hide sprites

with -recorder=gif:
s - Start recording to frame.gif
d - Stop recording

with -recorder=jpeg:
s - Save screenshot to frame.jpg
```

## Support

Audio has not been implemented yet.

Battery backed saves is implemented.

### Mappers

- NROM
- MMC1
- UNROM
- CNROM
- MMC3
- ANROM
- MMC2
