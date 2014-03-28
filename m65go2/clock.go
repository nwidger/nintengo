package m65go2

import (
	_ "fmt"
	"sync"
	"time"
)

// Represents a clock signal for an IC.  Once a Clock is started, it
// maintains a 'ticks' counters which is incremented at a specific
// interval.
type Clocker interface {
	// Returns the current value of the Clocker's ticks counter.
	Ticks() uint64

	// Starts the clock
	Start() (ticks uint64)

	// Stops the clock
	Stop()

	// Blocks the calling thread until the given tick has arrived.
	// Returns immediately if the clock has already passed the
	// given tick.
	Await(tick uint64) (ticks uint64)

	// Increment the Clocker's ticks counter by the given amount.
	Increment(amount uint64) (ticks uint64)
}

// Represents a basic clock that increments at specific intervals.
type Clock struct {
	rate     time.Duration
	ticks    uint64
	ticker   *time.Ticker
	stopChan chan int
	mutex    sync.Mutex
	waiting  map[uint64][]chan int
}

// Returns a pointer to a new Clock which increments its ticker at
// intervals of 'rate'.  The returned Clock has not been started and
// its ticks counter is zero.
func NewClock(rate time.Duration) *Clock {
	return &Clock{
		rate:     rate,
		ticks:    0,
		ticker:   nil,
		stopChan: make(chan int),
		waiting:  make(map[uint64][]chan int),
	}
}

func (clock *Clock) wakeWaiting() {
	if Ca, ok := clock.waiting[clock.ticks]; ok {
		for _, C := range Ca {
			C <- 1
		}

		delete(clock.waiting, clock.ticks)
	}
}

func (clock *Clock) maintainTime() {
	for {
		select {
		case <-clock.stopChan:
			clock.ticker.Stop()
			clock.ticker = nil
			return
		case _ = <-clock.ticker.C:
			clock.mutex.Lock()
			clock.ticks++
			clock.wakeWaiting()
			clock.mutex.Unlock()
		}
	}
}

func (clock *Clock) Ticks() (ticks uint64) {
	clock.mutex.Lock()
	ticks = clock.ticks
	clock.mutex.Unlock()

	return
}

func (clock *Clock) Start() (ticks uint64) {
	ticks = clock.ticks

	if clock.ticker == nil {
		clock.ticker = time.NewTicker(clock.rate)
		go clock.maintainTime()
	}

	return
}

func (clock *Clock) Stop() {
	if clock.ticker != nil {
		clock.stopChan <- 1
	}
}

func (clock *Clock) Increment(amount uint64) (ticks uint64) {
	clock.mutex.Lock()

	for i := uint64(1); i <= amount; i++ {
		clock.ticks++
		clock.wakeWaiting()
	}

	ticks = clock.ticks

	clock.mutex.Unlock()

	return
}

func (clock *Clock) Await(tick uint64) (ticks uint64) {
	clock.mutex.Lock()
	ticks = clock.ticks

	if ticks >= tick {
		clock.mutex.Unlock()
	} else {
		C := make(chan int, 1)
		clock.waiting[tick] = append(clock.waiting[tick], C)
		clock.mutex.Unlock()
		<-C
	}

	return
}

// Represents a clock divider which divides the tick frequency of
// another Clock so that it ticks at a slower rate.
type Divider struct {
	master  Clocker
	divisor uint64
}

// Returns a pointer to a new DividerCLock which divides the tick rate
// of 'master' Clocker by 'divisor'.
func NewDivider(master Clocker, divisor uint64) *Divider {
	return &Divider{divisor: divisor, master: master}
}

func (clock *Divider) Ticks() uint64 {
	return clock.master.Ticks() / clock.divisor
}

func (clock *Divider) Start() (ticks uint64) {
	return clock.master.Start() / clock.divisor
}

func (clock *Divider) Stop() {
	clock.master.Stop()
}

func (clock *Divider) Await(tick uint64) (ticks uint64) {
	return clock.master.Await(tick*clock.divisor) / clock.divisor
}

func (clock *Divider) Increment(amount uint64) (ticks uint64) {
	ticks = clock.master.Increment(amount * clock.divisor)
	return
}
