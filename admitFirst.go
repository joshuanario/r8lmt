package r8lmt

import "time"

//admit-first pipelining, https://blog.golang.org/pipelines

func NewAdmitFirstSpamChan (r8lmtChan chan interface{}, config *Config) chan interface{} {
	//inspired by leo lara https://disqus.com/by/disqus_BI7TGHPb0v/
	//first spamChan is admitted ("pass through") and begins the "reservation" (delay of time)
	//returned channel is a channel expecting spammy inputs
	//r8lmtChan is the ratelimited channel of the returned channel (spam of inputs)
	spamChan := make(chan interface{}) //pipelined channel for rx only, no need to close //todo add direction
	var buffer interface{}             //data received from spammy channel
	var ok bool
	passthru := true	//flag to admit a checkin after the end of reservation wait
	var timer *time.Timer
	admit := func(newdata interface{}) {
		go func() {
			r8lmtChan<- newdata
		}()
	}
	resetTimer := func() {
		timer = time.NewTimer(config.Reservation)
	}
	listenAndAdmitNextCheckin := func() {
		buffer, ok := <-spamChan	//blocks for a "check in"
		admit(buffer)	//admit a checkin
		// If channel closed exit goroutine
		if !ok {
			return
		}
		resetTimer()
	}
	go func() {
		defer close(r8lmtChan)
		listenAndAdmitNextCheckin() //wait for admitted spamChan then send to start next reservation wait
		for {
			//todo needs thorough review
			select {
			case <-timer.C:	//reservation wait is over
				if config.WillAdmitAfter { //admit a checkin at end of reservation
					admit(buffer)
					resetTimer() //after admit and reset, check if state is open for admission or not
					if passthru {
						listenAndAdmitNextCheckin() //next checkin is not waitlisted and quickly admitted
					} else {
						passthru = true	//reset admission flag and restart the reservation wait
						buffer = nil //reset buffer for next reservation
					}
				} else {	//discard buffer at end of reservation wait
					listenAndAdmitNextCheckin() //listen for checkin to be admitted
				}
			case buffer, ok = <-spamChan: //a checkin during the reservation wait
				if !ok {
					return
				}
				if config.IsReservationExpandable {	//debounce, extend the reservation wait by resetting the countdown
					timer.Reset(config.Reservation)
				}	//else do not extend the delay
				if config.WillAdmitAfter { //send latest checkin at end of delay a
					passthru = false	//mark next checkin as reserved
				}
			}// chan sel
		}// loop
	}()
	return spamChan
}