package nes

import (
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"os"
)

func loop(conn net.Conn, incoming chan<- Packet, outgoing <-chan Packet) (err error) {
	dec := gob.NewDecoder(conn)
	enc := gob.NewEncoder(conn)
	go func(c <-chan Packet, conn net.Conn, enc *gob.Encoder) {
		for {
			pkt := <-c
			err := enc.Encode(&pkt)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Encode err: %s\n", err)
				if err == io.EOF {
					break
				}
			}
		}
		conn.Close()
	}(outgoing, conn, enc)
	var pkt Packet
	for {
		err = dec.Decode(&pkt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Decode err: %s\n", err)
			if err == io.EOF {
				break
			}
		} else {
			incoming <- pkt
		}
	}
	conn.Close()
	return
}

type Bridge struct {
	nes      *NES
	incoming chan Packet
	outgoing chan Packet
	addr     string
	active   bool
}

func newBridge(nes *NES, addr string) (bridge *Bridge) {
	cin := make(chan Packet, 16)
	cout := make(chan Packet, 16)
	bridge = &Bridge{
		nes:      nes,
		incoming: cin,
		outgoing: cout,
		addr:     addr,
	}
	return
}

func (bridge *Bridge) runAsMaster() error {
	if len(bridge.addr) == 0 {
		return nil
	}

	ln, err := net.Listen("tcp", bridge.addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Listen error: %s\n", err)
		return err
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Accept error: %s\n", err)
			// Serve next conn
			continue
		}
		lock := <-bridge.nes.lock
		bridge.active = true
		ev, err := bridge.nes.getLoadStateEvent()
		bridge.outgoing <- Packet{
			Tick: bridge.nes.Tick,
			Ev:   ev,
		}
		bridge.nes.lock <- lock
		if err == nil {
			if err := loop(conn, bridge.incoming, bridge.outgoing); err != nil {
				fmt.Fprintf(os.Stderr, "Serving slave error: %s\n", err)
			}
		}
		bridge.active = false
	}
	return err
}

func (bridge *Bridge) runAsSlave() error {
	if len(bridge.addr) == 0 {
		return nil
	}

	conn, err := net.Dial("tcp", bridge.addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Connecting error: %s\n", err)
		return err
	}
	return loop(conn, bridge.incoming, bridge.outgoing)
}
