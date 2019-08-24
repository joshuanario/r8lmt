package r8lmt

import "time"

//AdmitFirstPipeline
//inspired by leo lara https://disqus.com/by/disqus_BI7TGHPb0v/
//first input is "admitted" ("pass through") and begins the "reservation" (delay of time)
//any inputs "checking in" during the reservation gets "wait listed" for admission at end of reservation
func AdmitFirstPipeline(rl *RateLimit) {
	spamChan := rl.Spammy  //pipelined channel for rx only, no need to close //todo add direction
	var buffer interface{} //data received from spammy channel
	var ok bool
	var timer *time.Timer
	admit := func(newdata interface{}) {
		go func() {
			rl.Limited <- newdata
		}()
	}
	resetTimer := func() {
		timer = time.NewTimer(rl.Reservation.Duration)
	}
	listenAndAdmitNextCheckin := func() {
		buffer, ok = <-spamChan //blocks for a "check in"
		admit(buffer)           //admit a checkin
		// If channel closed exit goroutine
		if !ok {
			return
		}
		resetTimer()
	}
	go func() {
		defer close(rl.Limited)
		listenAndAdmitNextCheckin() //wait for admitted spamChan then send to start next reservation wait
		for {
			select {
			case <-timer.C: //reservation wait is over
				listenAndAdmitNextCheckin() //listen for checkin to be admitted
			case buffer, ok = <-spamChan: //a checkin during the reservation wait
				if !ok {
					return
				}
				if rl.Reservation.IsExtensible { //debounce, extend the reservation wait by resetting the countdown
					timer.Reset(rl.Reservation.Duration)
				} //else do not extend the delay
			} // chan sel
		} // loop
	}()
}
