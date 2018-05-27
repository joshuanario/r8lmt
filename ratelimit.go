package r8lmt

import "time"

type RateLimit rune

type Reservation time.Duration

const (
	DEBOUNCE = RateLimit('d')
	THROTTLE = RateLimit('t')
)

type Config struct {
	WillAdmitAfter bool
	Reservation time.Duration
	IsReservationExpandable bool
}

func Debouncer (out chan<- interface{}, in <-chan interface{}, delay time.Duration,	leading bool, trailing bool) {
	goPipeline(out, in, delay, DEBOUNCE, leading, trailing)
}

func Throttler (out chan<- interface{}, in <-chan interface{}, delay time.Duration,	leading bool, trailing bool) {
	goPipeline(out,in, delay, THROTTLE, leading, trailing)
}

func goPipeline (out chan<- interface{}, in <-chan interface{}, reservation time.Duration, scheme RateLimit, leading bool, trailing bool) {
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
		Reservation:reservation,
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
				out<- p
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
				spamChan<- p
				if !ok {
					done<- true
					return
				}
			}
		}
	}()
	<-done // hang until done
	close(done)// todo check closing channels https://go101.org/article/channel-closing.html
	close(spamChan)// todo check closing channels https://go101.org/article/channel-closing.html
	close(out)// todo check closing channels https://go101.org/article/channel-closing.html
}