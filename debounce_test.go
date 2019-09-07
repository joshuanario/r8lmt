package r8lmt_test

import (
	"sync"
	"testing"
	"time"

	"github.com/joshuanario/r8lmt"
)

func TestLeadingDebouncer(t *testing.T) {
	out := make(chan interface{})
	in := make(chan interface{})
	dur := 1 * time.Microsecond
	r8lmt.Debouncer(out, in, dur, true)
	var wg sync.WaitGroup
	for c := 0; c < 256; c++ {
		wg.Add(1)
		go func(iifeC int) {
			in <- 257 - iifeC
			wg.Done()
		}(c)
	}
	wg.Wait()
	o, ok := <-out
	if !ok {
		t.Errorf("cannot receive from out channel")
	}
	casted := o.(int)
	if casted <= 0 || casted > 257 {
		t.Fatalf("non-zero expected; outcome %d", casted)
	}
	oo, ok := <-out
	if !ok {
		t.Errorf("cannot receive from out channel")
	}
	casted = oo.(int)
	if casted <= 0 || casted > 257 {
		t.Fatalf("non-zero expected; outcome %d", casted)
	}
	select {
	case oo, _ = <-out:
		casted := o.(int)
		t.Fatalf("expected empty channel; outcome %d", casted)
	case <-time.After(4 * dur):
	}
}

func TestNonleadingDebouncer(t *testing.T) {
	out := make(chan interface{})
	in := make(chan interface{})
	dur := 1 * time.Microsecond
	r8lmt.Debouncer(out, in, dur, false)
	var wg sync.WaitGroup
	for c := 0; c < 256; c++ {
		wg.Add(1)
		go func(iifeC int) {
			in <- 257 - iifeC
			wg.Done()
		}(c)
	}
	wg.Wait()
	o, ok := <-out
	if !ok {
		t.Errorf("cannot receive from out channel")
	}
	casted := o.(int)
	if casted <= 0 || casted > 257 {
		t.Fatalf("non-zero expected; outcome %d", casted)
	}
	o, _ = <-out
	casted = o.(int) //todo needs review
	select {
	case o, _ = <-out:
		casted = o.(int)
		t.Fatalf("expected empty channel; outcome %d", casted)
	case <-time.After(4 * dur):
	}
}
