nintengo
========

[![Build Status](https://travis-ci.org/nwidger/nintengo.svg?branch=master)](https://travis-ci.org/nwidger/nintengo)

An NES emulator written in Go

![Super Mario Bros.](http://i.imgur.com/g6ogqv7.gif "Super Mario Bros.")

![Donkey Kong](http://i.imgur.com/0SIbydD.gif "Donkey Kong")

![Excitebike](http://i.imgur.com/NTYlltB.gif "Excitebike")

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

NROM (mapper 0) has been implemented, no other mappers are supported yet.

Audio has not been implemented yet.

Battern backed saves is not implemented yet.
