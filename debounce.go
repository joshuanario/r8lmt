package r8lmt

import "time"

//Debouncer - wrapper for NewLimiter
func Debouncer(out chan interface{}, in chan interface{}, delay time.Duration, leading bool, trailing bool) {
	bw := RESERVEFIRST
	if leading {
		bw = ADMITFIRST
	}
	rl := NewLimiter(in, out, delay, DEBOUNCE, bw)
	startPipeline(*rl)
}
