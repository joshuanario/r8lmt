package r8lmt

import "time"

type Style rune

const (
	DEBOUNCE = Style('d')
	THROTTLE = Style('t')
)

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

type Config struct { // todo nuke config ifo rl
	WillAdmitAfter bool //todo maxReservations
	Reservation    time.Duration
	IsExtensible   bool
}

func startPipeline(rl RateLimit) {
	if rl.Reservation.BeforeWait == ADMITFIRST {
		AdmitFirstPipeline(&rl)
		ReserveFirstPipeline(&rl)
	}
}
