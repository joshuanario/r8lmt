package r8lmt

import "time"

type Config struct {
	WillAdmitAfter bool
	Reservation time.Duration
	IsReservationExpandable bool
}

// todo check closing channels https://go101.org/article/channel-closing.html

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

func NewAdmitFirstSpamChan (r8lmtChan chan interface{}, config *Config) chan interface{} {
	//inspired by leo lara https://disqus.com/by/disqus_BI7TGHPb0v/
	//first input is admitted ("pass through") and begins the "reservation" (delay of time)
	//returned channel is a channel expecting spammy inputs
	//r8lmtChan is the ratelimited channel of the returned channel (spam of inputs)
	input := make(chan interface{})	//synchronous channel for pipelining
	var buffer interface{}	//data received from spammy channel
	var ok bool
	passthru := true	//flag to admit an input after the end of reservation wait
	var timer *time.Timer
	sendbuf := func() {
		go func() {
			r8lmtChan <-buffer
		}()
		timer = time.NewTimer(config.Reservation)
	}
	waitpassthruinput := func() {
		buffer, ok := <-input
		// If channel closed exit goroutine
		if !ok {
			go func() {
				r8lmtChan <-buffer
			}()
			return
		}
		sendbuf()
	}
	go func() {
		defer close(r8lmtChan)
		waitpassthruinput()	//wait for admitted input then send to start next reservation wait
		for {
			select {
			case <-timer.C:	//reservation wait is over
				if config.WillAdmitAfter { //send buffer at end of reservation
					sendbuf()	//after send, check if input is admitted or reserved
					if passthru {
						waitpassthruinput()	//next input is not reserved and admitted to r8lmtChan
					} else {
						passthru = true	//reset admission flag and restart the reservation wait
						buffer = nil //reset buffer for next reservation
					}
				} else {	//discard buffer at end of reservation wait
					waitpassthruinput() //just wait for next admitted input
				}
			case buffer, ok = <-input:	//data was received from input during the reservation wait
				if !ok {
					return
				}
				if config.IsReservationExpandable {	//extend the reservation wait by restarting the wait countdown
					timer.Reset(config.Reservation)
				}	//else do not extend the delay
				if config.WillAdmitAfter { //send latest input at end of delay a
					passthru = false	//mark next input as reserved
				}
			}// chan sel
		}// loop
	}()
	return input
}