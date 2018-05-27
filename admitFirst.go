package r8lmt

import "time"

func NewAdmitFirstSpamChan (r8lmtChan chan interface{}, config *Config) chan interface{} {
	//inspired by leo lara https://disqus.com/by/disqus_BI7TGHPb0v/
	//first spamChan is admitted ("pass through") and begins the "reservation" (delay of time)
	//returned channel is a channel expecting spammy inputs
	//r8lmtChan is the ratelimited channel of the returned channel (spam of inputs)
	spamChan := make(chan interface{}) //synchronous channel for pipelining
	var buffer interface{}             //data received from spammy channel
	var ok bool
	passthru := true	//flag to admit an spamChan after the end of reservation wait
	var timer *time.Timer
	sendbuf := func() {
		go func() {
			r8lmtChan <-buffer
		}()
		timer = time.NewTimer(config.Reservation)
	}
	waitpassthruinput := func() {
		buffer, ok := <-spamChan
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
		waitpassthruinput()	//wait for admitted spamChan then send to start next reservation wait
		for {
			select {
			case <-timer.C:	//reservation wait is over
				if config.WillAdmitAfter { //send buffer at end of reservation
					sendbuf()	//after send, check if spamChan is admitted or reserved
					if passthru {
						waitpassthruinput()	//next spamChan is not reserved and admitted to r8lmtChan
					} else {
						passthru = true	//reset admission flag and restart the reservation wait
						buffer = nil //reset buffer for next reservation
					}
				} else {	//discard buffer at end of reservation wait
					waitpassthruinput() //just wait for next admitted spamChan
				}
			case buffer, ok = <-spamChan: //data was received from spamChan during the reservation wait
				if !ok {
					return
				}
				if config.IsReservationExpandable {	//extend the reservation wait by restarting the wait countdown
					timer.Reset(config.Reservation)
				}	//else do not extend the delay
				if config.WillAdmitAfter { //send latest spamChan at end of delay a
					passthru = false	//mark next spamChan as reserved
				}
			}// chan sel
		}// loop
	}()
	return spamChan
}