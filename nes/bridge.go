package nes

import (
	"encoding/gob"
	"fmt"
	"io"
	"net"
)

func loop(conn net.Conn, incoming chan<- Packet, outgoing <-chan Packet) error {
	fmt.Println("Start loop")
	dec := gob.NewDecoder(conn)
	enc := gob.NewEncoder(conn)
	go func(c <-chan Packet, conn net.Conn, enc *gob.Encoder) {
		for {
			pkt := <-c
			// fmt.Println("Encoding: ", pkt)
			err := enc.Encode(&pkt)
			if err != nil {
				fmt.Println("Error Encoding: ", err)
				break
			}
		}
		conn.Close()
	}(outgoing, conn, enc)
	var pkt Packet
	for {
		//fmt.Println("Decoding..")
		err := dec.Decode(&pkt)
		if err != nil {
			fmt.Println("Decode err: ", err)
			if err == io.EOF {
				break
			}
			return err
		} else {
			//fmt.Println("Decoded: ", pkt)
			incoming <- pkt
		}
	}
	conn.Close()
	return nil
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
	ln, err := net.Listen("tcp", bridge.addr)
	if err != nil {
		fmt.Println(err)
		return err
	}
	for {
		conn, err := ln.Accept()
		fmt.Println("Got conn")
		if err != nil {
			fmt.Println("Error: ", err)
			break
		}
		fmt.Println("Getting LoadStateEvent")
		lock := <-bridge.nes.lock
		bridge.active = true
		ev, err := bridge.nes.getLoadStateEvent()
		bridge.outgoing <- Packet{
			Tick: bridge.nes.Tick,
			Ev:   ev,
		}
		bridge.nes.lock <- lock
		fmt.Println("Startinig loop")
		if err == nil {
			loop(conn, bridge.incoming, bridge.outgoing)
		}
		bridge.active = false
	}
	return nil
}

func (bridge *Bridge) runAsSlave() error {
	conn, err := net.Dial("tcp", bridge.addr)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return loop(conn, bridge.incoming, bridge.outgoing)
}
