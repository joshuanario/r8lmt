package r8lmt

import "time"

//AdmitFirstPipeline
//inspired by leo lara https://disqus.com/by/disqus_BI7TGHPb0v/
//first input is "admitted" ("pass through") and begins the "reservation" (delay of time)
//any inputs "checking in" during the reservation gets "wait listed" for admission at end of reservation
func AdmitFirstPipeline(rl *RateLimit) {
	spamChan := rl.Spammy //todo add direction
	var buffer interface{}
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
		buffer, ok = <-spamChan
		admit(buffer)
		if !ok {
			return
		}
		resetTimer()
	}
	go func() {
		defer close(rl.Limited)
		listenAndAdmitNextCheckin()
		for {
			select {
			case <-timer.C:
				listenAndAdmitNextCheckin()
			case buffer, ok = <-spamChan:
				if !ok {
					return
				}
				if rl.Reservation.IsExtensible {
					timer.Reset(rl.Reservation.Duration)
				}
			}
		}
	}()
}
