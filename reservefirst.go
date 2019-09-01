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
	reserved := false
	lmtd := out
	admit := func(newdata interface{}) {
		go func() {
			lmtd <- newdata
		}()
	}

	go func() {
		timer := time.NewTimer(rl.Reservation.Duration)
		for {
			select {
			case spam, ok := <-spmy:
				if !reserved {
					buffer = spam
					reserved = true
				}
				if !ok {
					return
				}
				if rl.Reservation.IsExtensible {
					timer.Reset(rl.Reservation.Duration)
				}
			case <-timer.C:
				if reserved {
					admit(buffer)
				}
				timer = time.NewTimer(rl.Reservation.Duration)
				reserved = false
			}
		}
	}()
}
