package r8lmt

import "time"

type RateLimit rune

const (
	DEBOUNCE = 'd'
	THROTTLE = 't'
)

func goPipeline (out chan<- interface{}, in <-chan interface{}, delay time.Duration, scheme RateLimit, leading bool, trailing bool) {
	var isReservationExpandable bool
	switch scheme {
	case THROTTLE:
		isReservationExpandable = false
	case DEBOUNCE:
	default:
		isReservationExpandable = true
	}
	done := make(chan bool)
	r8lmtChan := make(chan interface{})
	spamChan := make(chan interface{})
	config := &Config{
		Reservation:delay,
		IsReservationExpandable:isReservationExpandable,
		WillAdmitAfter:trailing,
	}
	if leading {
		spamChan = NewAdmitFirstSpamChan(r8lmtChan, config)
	} else {
		spamChan = NewReserveFirstSpamChan(r8lmtChan, config)
	}
	go func() {
		//go routine to rx from r8lmt then send to out
		for {
			select{
			case p, ok := <-r8lmtChan:
				out <- p
				if !ok {
					done<-true
					return
				}
			case <-done:
				return
			}
		}
	}()
	go func() {
		//go routine to pass in to spam
		for {
			select{
			case <-done:
				return
			case p, ok := <-in:
				spamChan <- p
				if !ok {
					done<-true
					return
				}
			}
		}
	}()
	<-done // hang until done
	close(done)
	close(spamChan)
	close(out)
}