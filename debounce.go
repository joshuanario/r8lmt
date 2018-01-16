package r8lmt

import "time"

func Debouncer (out chan<- interface{}, in <-chan interface{}, delay time.Duration,	leading bool, trailing bool) {
	goPipeline(out, in, delay, DEBOUNCE, leading, trailing)
}