package r8lmt

import "time"

//Throttler - wrapper for NewLimiter
func Throttler(out chan interface{}, in chan interface{}, delay time.Duration, leading bool) {
	bw := RESERVEFIRST
	if leading {
		bw = ADMITFIRST
	}
	rl := NewLimiter(in, out, delay, THROTTLE, bw)
	startPipeline(*rl)
}
