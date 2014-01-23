package nintengo

import (
	"testing"
)

func TestMario(t *testing.T) {
	t.Skip()

	nes, err := NewNES("/Users/nwidger/Desktop/Super Mario Bros.nes")

	if err != nil {
		t.Errorf("Error loading valid Rom: %v\n", err)
	}

	nes.cpu.EnableDecode()
	nes.Reset()
	nes.Run()
	return
}
