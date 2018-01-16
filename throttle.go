package r8lmt

import "time"

type Throttle struct {
	delayfirst bool
	repetitive bool
	delay time.Duration
	throttled chan interface{}
}

func Throttler (out chan<- interface{}, in <-chan interface{}, delay time.Duration,	leading bool, trailing bool) {
	goPipeline(out,in, delay, THROTTLE, leading, trailing)
}