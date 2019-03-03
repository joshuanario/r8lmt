package r8lmt

import "time"

func Throttler(out chan interface{}, in chan interface{}, delay time.Duration, leading bool, trailing bool) {
	bw := RESERVEFIRST
	if leading {
		bw = ADMITFIRST
	}
	rl := NewLimiter(in, out, delay, THROTTLE, bw)
	startPipeline(*rl)
}
