package r8lmt

import "time"

//reserve-first pipelining, https://blog.golang.org/pipelines

func NewReserveFirstSpamChan (r8lmtChan chan interface{}, config *Config) chan interface{} {
	//Caleb Lloyd's idea of rate limiting, first input is admitted ("debounced") to the end of "reservation wait",
	//an output is expected at end of a reservation irregardless of subsequent inputs during the reservation wait
	//another mental model for this is: the first input to "check in" gets wait listed for a certain time ("reservation
	// wait") before admission and any other inputs that attempt to check in while someone is wait listed gets rejected
	//returned channel is a channel expecting spammy inputs
	//r8lmtChan is the ratelimited channel of the returned channel (spam of inputs)
	//Note that WillAdmitAfter is irrelevant for a reserve-first ratelimiting
	spamChan := make(chan interface{}) //pipelined channel for rx only, no need to close //todo add direction
	var buffer interface{}
	var ok bool
	unreserved := false	//flag when something is wait listed
	throttledPipeline := func() {
		defer close(r8lmtChan)	//todo fix channel closing
		timer := time.NewTimer(config.Reservation)
		for {
			select {
			case buffer, ok = <-spamChan:
				if !ok {
					// If channel closed exit goroutine
					return
				}
				timer.Reset(config.Reservation)
				unreserved = true
			case <-timer.C:
				go func() {
					r8lmtChan<- buffer
				}()
				if unreserved {
					buffer, ok = <-spamChan
					if !ok {
						return
					}
					timer = time.NewTimer(config.Reservation)
					unreserved = false
				}
			}
		}
	}
	debouncedPipeline := func() {
		defer close(r8lmtChan)	//todo fix channel closing
		for {
			select {
			case buffer, ok = <-spamChan:
				if !ok {
					return
				}
				unreserved = true
			case <-time.After(config.Reservation):
				go func() {
					r8lmtChan<- buffer
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

	if config.IsReservationExpandable {	//aka debouncer where "reservation wait" interval can be extended due to spam
		go throttledPipeline()
	} else {	//aka fixed-rate throttler where reservation intervals are consistent
		go debouncedPipeline()
	}
	return spamChan
}
