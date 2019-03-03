package r8lmt

import "time"

func Debouncer(out chan interface{}, in chan interface{}, delay time.Duration, leading bool, trailing bool) {
	//(out, in, delay, DEBOUNCE, leading, trailing)
	bw := RESERVEFIRST
	if leading {
		bw = ADMITFIRST
	}
	rl := NewLimiter(in, out, delay, DEBOUNCE, bw)
	startPipeline(*rl)
}
