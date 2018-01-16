package r8lmt

import "time"

func Throttler (out chan<- interface{}, in <-chan interface{}, delay time.Duration,	leading bool, trailing bool) {
	goPipeline(out,in, delay, THROTTLE, leading, trailing)
}