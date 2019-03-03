package r8lmt

import "time"

type BeforeWait int

const (
	RESERVEFIRST = BeforeWait('r')
	ADMITFIRST   = BeforeWait('a')
)

type AfterWait int

const (
	CLEARALL = AfterWait('c')
	KEEPALL  = AfterWait('k')
)

type OnWait int

const (
	FIRSTINLINE = OnWait('f')
	LASTINLINE  = OnWait('l')
)

type RateLimit struct {
	MaxReservations int
	Spammy          chan interface{}
	Limited         chan interface{}
	Reservation     Reservation
	WaitList        WaitList
}

type Reservation struct {
	BeforeWait   BeforeWait
	AfterWait    AfterWait
	IsExtensible bool
	Duration     time.Duration
}

type WaitList struct {
	OnWait  OnWait
	Maximum int
}

func NewWaitList() WaitList {
	return struct {
		OnWait  OnWait
		Maximum int
	}{OnWait: FIRSTINLINE, Maximum: 0}
}

func NewLimiter(in chan interface{}, out chan interface{}, t time.Duration, s Style, bw BeforeWait) *RateLimit {
	var ext bool = false
	if s == DEBOUNCE {
		ext = true
	}
	r := Reservation{
		AfterWait:    CLEARALL,
		BeforeWait:   bw,
		Duration:     t,
		IsExtensible: ext,
	}
	ret := RateLimit{WaitList: NewWaitList(), Reservation: r, Limited: out, Spammy: in, MaxReservations: 0}
	return &ret
}

func startPipeline(rl RateLimit) {
	done := make(chan bool)
	r8lmtChan := make(chan interface{})
	spamChan := make(chan interface{})
	config := &Config{ // todo nuke config ifo rl
		Reservation:    rl.Reservation.Duration,
		IsExtensible:   rl.Reservation.IsExtensible,
		WillAdmitAfter: rl.WaitList.Maximum > 0,
	}
	if rl.Reservation.BeforeWait == ADMITFIRST {
		spamChan = NewAdmitFirstSpamChan(r8lmtChan, config)
	} else {
		spamChan = NewReserveFirstSpamChan(r8lmtChan, config)
	}
	go func() {
		//go routine to rx from r8lmt then send to Limited
		for {
			select {
			case p, ok := <-r8lmtChan:
				rl.Limited <- p
				if !ok {
					done <- true
					return
				}
			case <-done:
				return
			}
		}
	}()
	go func() {
		//go routine to pass Spammy to spam
		for {
			select {
			case <-done:
				return
			case p, ok := <-rl.Spammy:
				spamChan <- p
				if !ok {
					done <- true
					return
				}
			}
		}
	}()
	<-done            // hang until done
	close(done)       // todo check closing channels https://go101.org/article/channel-closing.html
	close(spamChan)   // todo check closing channels https://go101.org/article/channel-closing.html
	close(rl.Limited) // todo check closing channels https://go101.org/article/channel-closing.html
}
