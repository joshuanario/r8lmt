package r8lmt

import "time"

//AdmitFirstPipeline -
//inspired by leo lara https://disqus.com/by/disqus_BI7TGHPb0v/
//first input is "admitted" ("pass through") and begins the "reservation" (delay of time)
//any inputs "checking in" during the reservation gets "wait listed" for admission at end of reservation
func AdmitFirstPipeline(rl *RateLimit, out chan<- interface{}, in <-chan interface{}) {
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
	go func() {
		lmtd := out
		var buffer interface{}
		var ok bool
		var timer *time.Timer
		admit := func(newdata interface{}) {
			go func() {
				lmtd <- newdata
			}()
		}
		listenAndAdmitNextCheckin := func() {
			buffer, ok = <-spmy
			admit(buffer)
			if !ok {
				return
			}
			timer = time.NewTimer(rl.Reservation.Duration)
		}
		listenAndAdmitNextCheckin()
		for {
			select {
			case <-timer.C:
				listenAndAdmitNextCheckin()
			case buffer, ok = <-spmy:
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
