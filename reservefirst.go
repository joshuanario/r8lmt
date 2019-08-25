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
	unreserved := false
	lmtd := out
	// admit := func(newdata interface{}) {
	// 	go func() {
	// 		lmtd <- newdata
	// 	}()
	// }
	if rl.Reservation.IsExtensible {
		go func() {
			timer := time.NewTimer(rl.Reservation.Duration)
			for {
				select {
				case spam, ok := <-spmy:
					if unreserved {
						buffer = spam
					}
					if !ok {
						return
					}
					timer.Reset(rl.Reservation.Duration)
					unreserved = true
				case <-timer.C:
					lmtd <- buffer
					timer = time.NewTimer(rl.Reservation.Duration)
					unreserved = false
				}
			}
		}()
	} else {
		go func() {
			for {
				select {
				case spam, ok := <-spmy:
					if unreserved {
						buffer = spam
					}
					if !ok {
						return
					}
					unreserved = true
				case <-time.After(rl.Reservation.Duration):
					lmtd <- buffer
					unreserved = false
				}
			}
		}()
	}
}
