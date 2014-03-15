package nes

import (
	"testing"
	"time"
)

func TestMario(t *testing.T) {
	t.Skip()

	nes, err := NewNES("Super Mario Bros.nes", nil)

	if err != nil {
		t.Errorf("Error loading valid Rom: %v\n", err)
	}

	// nes.cpu.EnableDecode()
	nes.Reset()
	go nes.Run()

	<-time.After(time.Second * 30)

	return
}
