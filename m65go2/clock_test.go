package m65go2

import (
	"testing"
	"time"
)

func BenchmarkAwait(b *testing.B) {
	rate, _ := time.ParseDuration("552ns")
	clock := NewClock(rate)
	ticks := clock.Ticks()
	go clock.Start()

	b.ResetTimer()
	clock.Await(ticks + 1)
}

func TestAwait(t *testing.T) {
	t.Skip()

	rate, _ := time.ParseDuration("46ns")
	clock := NewClock(rate)
	ticks := clock.Ticks()

	go clock.Start()
	newTicks := clock.Await(ticks + 100)

	if newTicks != ticks+100 {
		t.Error("Ticks did not increment by 100!")
	}

	clock.Stop()
}
