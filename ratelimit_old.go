package r8lmt

import "time"

type Style rune

const (
	DEBOUNCE = Style('d')
	THROTTLE = Style('t')
)

type Config struct {
	WillAdmitAfter bool //todo maxReservations
	Reservation    time.Duration
	IsExtensible   bool
}

//func goPipeline(out chan<- interface{}, in <-chan interface{}, reservation time.Duration, style Style, leading bool, trailing bool) {
//	var isReservationExtensible bool
//	switch style {
//	case THROTTLE:
//		isReservationExtensible = false
//	case DEBOUNCE:
//	default:
//		isReservationExtensible = true
//	}
//	done := make(chan bool)
//	r8lmtChan := make(chan interface{})
//	spamChan := make(chan interface{})
//	config := &Config{
//		Reservation:    reservation,
//		IsExtensible:   isReservationExtensible,
//		WillAdmitAfter: trailing,
//	}
//	if leading {
//		spamChan = NewAdmitFirstSpamChan(r8lmtChan, config)
//	} else {
//		spamChan = NewReserveFirstSpamChan(r8lmtChan, config)
//	}
//	go func() {
//		//go routine to rx from r8lmt then send to out
//		for {
//			select {
//			case p, ok := <-r8lmtChan:
//				out <- p
//				if !ok {
//					done <- true
//					return
//				}
//			case <-done:
//				return
//			}
//		}
//	}()
//	go func() {
//		//go routine to pass in to spam
//		for {
//			select {
//			case <-done:
//				return
//			case p, ok := <-in:
//				spamChan <- p
//				if !ok {
//					done <- true
//					return
//				}
//			}
//		}
//	}()
//	<-done          // hang until done
//	close(done)     // todo check closing channels https://go101.org/article/channel-closing.html
//	close(spamChan) // todo check closing channels https://go101.org/article/channel-closing.html
//	close(out)      // todo check closing channels https://go101.org/article/channel-closing.html
//}
