package r8lmt

import "time"

type Style rune

const (
	DEBOUNCE = Style('d')
	THROTTLE = Style('t')
)

type BeforeWait rune

const (
	RESERVEFIRST = BeforeWait('r')
	ADMITFIRST   = BeforeWait('a')
)

type AfterWait rune

const (
	CLEARALL = AfterWait('c')
	KEEPALL  = AfterWait('k')
)

type OnWait rune

const (
	FIRSTINLINE = OnWait('f')
	LASTINLINE  = OnWait('l')
)

type RateLimit struct {
	MaxReservations int //todo maxReservations
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

func newWaitList() *WaitList {
	return &WaitList{OnWait: FIRSTINLINE, Maximum: 0}
}

func NewLimiter(out chan<- interface{}, in <-chan interface{}, t time.Duration, s Style, bw BeforeWait) *RateLimit {
	ext := false
	if s == DEBOUNCE {
		ext = true
	}
	r := Reservation{
		AfterWait:    CLEARALL,
		BeforeWait:   bw,
		Duration:     t,
		IsExtensible: ext,
	}
	ret := RateLimit{WaitList: *newWaitList(), Reservation: r, MaxReservations: 0}
	startPipeline(ret, out, in)
	return &ret
}

func startPipeline(rl RateLimit, out chan<- interface{}, in <-chan interface{}) {
	if rl.Reservation.BeforeWait == ADMITFIRST {
		AdmitFirstPipeline(&rl, out, in)
	} else {
		ReserveFirstPipeline(&rl, out, in)
	}
}
