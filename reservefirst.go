package r8lmt

import "time"

func ReserveFirstPipeline(rl *RateLimit) {
	//Caleb Lloyd's idea of rate limiting, first input is admitted ("debounced") to the end of "reservation wait",
	//an output is expected at end of a reservation irregardless of subsequent inputs during the reservation wait
	//another mental model for this is: the first input to "check in" gets wait listed for a certain time ("reservation
	// wait") before admission and any other inputs that attempt to check in while someone is wait listed gets rejected
	// (wait listed at the back of the line)
	//Note that WillAdmitAfter is irrelevant for a reserve-first ratelimiting
	spamChan := rl.Spammy //pipelined channel for rx only, no need to close //todo add direction
	var buffer interface{}
	var ok bool
	unreserved := false //flag when something is wait listed
	debouncedPipeline := func() {
		defer close(rl.Limited) //closing channels https://go101.org/article/channel-closing.html
		timer := time.NewTimer(rl.Reservation.Duration)
		for {
			select {
			case buffer, ok = <-spamChan:
				if !ok {
					// If channel closed exit goroutine
					return
				}
				timer.Reset(rl.Reservation.Duration)
				unreserved = true
			case <-timer.C:
				go func() {
					rl.Limited <- buffer
				}()
				if unreserved {
					buffer, ok = <-spamChan
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
		defer close(rl.Limited) //closing channels https://go101.org/article/channel-closing.html
		for {
			select {
			case buffer, ok = <-spamChan:
				if !ok {
					return
				}
				unreserved = true
			case <-time.After(rl.Reservation.Duration):
				go func() {
					rl.Limited <- buffer
				}()
				if unreserved {
					buffer, ok = <-spamChan
					if !ok {
						// If channel closed exit goroutine
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
