package r8lmt_test

import (
	"fmt"
	"strconv"
	"time"

	"github.com/joshuanario/r8lmt"
)

func ExampleNewLimiter() {
	in := make(chan interface{})
	out := make(chan interface{})
	t := time.Second
	s := r8lmt.THROTTLE
	bw := r8lmt.RESERVEFIRST
	lmtr := r8lmt.NewLimiter(out, in, t, s, bw)
	str := strconv.FormatBool(lmtr.Reservation.IsExtensible)
	fmt.Printf("IsExtensible: %s\n", str)
	str = lmtr.Reservation.Duration.String()
	fmt.Printf("Duration: %s\n", str)
	go func() {
		time.Sleep(100 * time.Millisecond)
		in <- struct {
			text string
		}{
			text: "foobarfoobar",
		}
	}()
	foo := <-out
	stim, ok := foo.(struct {
		text string
	})
	if !ok {
		panic("Type assertion failed")
	}
	str = stim.text
	fmt.Printf("Stimulus: %s\n", str)
	// Output: IsExtensible: false
	// Duration: 1s
	// Stimulus: foobarfoobar
}
