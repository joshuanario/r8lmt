package r8lmt

import "time"

//Debouncer - wrapper for NewLimiter
func Debouncer(out chan<- interface{}, in <-chan interface{}, delay time.Duration, leading bool) {
	bw := RESERVEFIRST
	if leading {
		bw = ADMITFIRST
	}
	rl := NewLimiter(out, in, delay, DEBOUNCE, bw)
	startPipeline(*rl, out, in)
}
