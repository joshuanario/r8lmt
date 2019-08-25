package r8lmt

import "time"

//ReserveFirstPipeline
// Caleb Lloyd's idea of rate limiting, first input is admitted ("debounced") to the end of "reservation wait",
//an output is expected at end of a reservation irregardless of subsequent inputs during the reservation wait
//another mental model for this is: the first input to "check in" gets wait listed for a certain time ("reservation
//wait") before admission and any other inputs that attempt to check in while someone is wait listed gets rejected
//(wait listed at the back of the line)
//Note that WillAdmitAfter is irrelevant for a reserve-first ratelimiting
func ReserveFirstPipeline(rl *RateLimit, out chan<- interface{}, in <-chan interface{}) {
	spmy := make(chan interface{})
	go func() {
		defer close(spmy)
		for {
			buffer, ok := <-in
			spmy <- buffer
			if !ok {
				return
			}
		}
	}()
	var buffer interface{}
	var ok bool
	unreserved := false
	lmtd := out
	admit := func(newdata interface{}) {
		go func() {
			lmtd <- newdata
		}()
	}
	debouncedPipeline := func() {
		timer := time.NewTimer(rl.Reservation.Duration)
		for {
			select {
			case buffer, ok = <-spmy:
				if !ok {
					return
				}
				timer.Reset(rl.Reservation.Duration)
				unreserved = true
			case <-timer.C:
				admit(buffer)
				if unreserved {
					buffer, ok = <-spmy
					if !ok {
						return
					}
					timer = time.NewTimer(rl.Reservation.Duration)
					unreserved = false
				}
			}
		}
	}
	throttledPipeline := func() {
		for {
			select {
			case buffer, ok = <-spmy:
				if !ok {
					return
				}
				unreserved = true
			case <-time.After(rl.Reservation.Duration):
				admit(buffer)
				if unreserved {
					buffer, ok = <-spmy
					if !ok {
						return
					}
					unreserved = false
				}
			}
		}
	}

	if rl.Reservation.IsExtensible { //aka debouncer where "reservation wait" interval can be extended due to spam
		go debouncedPipeline()
	} else { //aka fixed-rate throttler where reservation intervals are consistent
		go throttledPipeline()
	}
}
