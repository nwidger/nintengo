package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"flag"

	"github.com/mitchellh/go-homedir"
	"github.com/nwidger/nintengo/http"
	"github.com/nwidger/nintengo/nes"
	"gopkg.in/yaml.v2"
)

func LoadConfig(options *nes.Options, filename string) (err error) {
	var buf []byte

	if buf, err = ioutil.ReadFile(filename); err != nil {
		return
	}

	if err = yaml.Unmarshal(buf, options); err != nil {
		fmt.Printf("*** Error loading config: %s\n", err)
		return
	}

	return
}

func main() {
	filename := ""

	options := &nes.Options{}

	flag.BoolVar(&options.CPUDecode, "cpu-decode", false, "decode CPU instructions")
	flag.StringVar(&options.Recorder, "recorder", "", "recorder to use: none | jpeg | gif")
	flag.StringVar(&options.AudioRecorder, "audio-recorder", "", "recorder to use: none | wav")
	flag.StringVar(&options.CPUProfile, "cpu-profile", "", "write CPU profile to file")
	flag.StringVar(&options.MemProfile, "mem-profile", "", "write memory profile to file")
	flag.StringVar(&options.HTTPAddress, "http", "", "HTTP service address (e.g., ':6060')")
	flag.StringVar(&options.Listen, "l", "", "Listen at address as master (e.g., ':8080')")
	flag.StringVar(&options.Connect, "c", "", "Connect to address as slave, <rom-file> will be ignored (e.g., 'localhost:8080')")
	flag.Parse()

	filename, err := homedir.Expand("~/.nintengorc")

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	} else {
		if _, err = os.Stat(filename); !os.IsNotExist(err) {
			if err = LoadConfig(options, filename); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
			}
		}
	}

	if len(flag.Args()) != 1 {
		if len(options.Connect) == 0 {
			fmt.Fprintf(os.Stderr, "usage: <rom-file>\n")
			return
		}
	} else {
		filename = flag.Arg(0)
	}

	nes, err := nes.NewNES(filename, options)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

	if options.HTTPAddress != "" {
		neserv := http.NewNEServer(nes, options.HTTPAddress)
		fmt.Println(options.HTTPAddress)
		go neserv.Run()
	}

	err = nes.Run()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
}
