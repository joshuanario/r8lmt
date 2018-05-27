package r8lmt

import "time"

func NewReserveFirstSpamChan (r8lmtChan chan interface{}, config *Config) chan interface{} {
	//Caleb Lloyd's idea of rate limiting, first input is reserved ("debounced") to the end of "reservation wait",
	//an output is expected at end of a reservation irregardless of subsequent inputs during the reservation wait
	//returned channel is a channel expecting spammy inputs
	//r8lmtChan is the ratelimited channel of the returned channel (spam of inputs)
	//Note that WillAdmitAfter is irrelevant for a reserve-first ratelimiting
	input := make(chan interface{})	//synchronous channel for pipelining
	var buffer interface{}
	var ok bool
	passthru := false
	sendbuf := func(waitinput func()) {
		go func() {
			r8lmtChan<- buffer
		}()
		if passthru{
			waitinput()
			passthru = false
		}
	}

	if config.IsReservationExpandable {	//aka debouncer where "reservation wait" interval can be extended due to spam
		go func() {
			defer close(r8lmtChan)
			timer := time.NewTimer(config.Reservation)
			aftersend := func() {
				buffer, ok = <-input
				if !ok {
					return
				}
				timer = time.NewTimer(config.Reservation)
			}
			for {
				select {
				case buffer, ok = <-input:
					if !ok {
						// If channel closed exit goroutine
						return
					}
					timer.Reset(config.Reservation)
					passthru = true
				case <-timer.C:
					sendbuf(aftersend)
				}
			}
		}()
		return input
	} else {	//aka fixed-rate throttler where reservation intervals are consistent
		go func() {
			defer close(r8lmtChan)
			aftersend := func() {
				buffer, ok = <-input
				if !ok {
					// If channel closed exit goroutine
					return
				}
			}
			for {
				select {
				case buffer, ok = <-input:
					if !ok {
						return
					}
					passthru = true
				case <-time.After(config.Reservation):
					sendbuf(aftersend)
				}
			}
		}()
		return input
	}
}
