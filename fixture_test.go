package r8lmt

import (
	"testing"
	"time"
)

type TestConfig struct {
	ReservationPulses int
	PulsePeriod       time.Duration
	MaxPulses         int
	TimeOut           time.Duration
}

const defaultTimeOut = time.Minute

const defaultPulsePeriod = time.Microsecond * 50

func newTestConfig(reservationPulses int, pulsePeriod time.Duration, maxPulses int, timeOut time.Duration) *TestConfig {
	return &TestConfig{
		ReservationPulses: reservationPulses,
		MaxPulses:         maxPulses,
		PulsePeriod:       pulsePeriod,
		TimeOut:           timeOut,
	}
}

func newTestChannels() *TestChannels { //todo better test chan
	//poller and spammer should have their own ready channel to fixture
	//then both chans has to be live for fixture to say GONOW
	//currently, poller and spammer are racing
	return &TestChannels{
		ReadySetGo: make(chan uint8),
		Finished:   make(chan bool),
	}
}

type TestFixture struct {
	dampchan     chan interface{}
	spamchan     chan interface{}
	channels     *TestChannels
	config       *TestConfig
	stimuli      Record
	expectations Record
}

func NewTestFixture(stimuli map[int]uint8, reservationPulses int, maxPulses int, reserveFirst bool, extensibleReservation bool, repetitive bool, expectations map[int]uint8) *TestFixture {
	tcnf := newTestConfig(reservationPulses, defaultPulsePeriod, maxPulses, defaultTimeOut)
	bw := ADMITFIRST
	if reserveFirst {
		bw = RESERVEFIRST
	}
	style := THROTTLE
	if extensibleReservation {
		style = DEBOUNCE
	}
	t := time.Duration(tcnf.ReservationPulses) * time.Duration(tcnf.PulsePeriod)
	limitchan := make(chan interface{})
	spamchan := make(chan interface{})
	sut := NewLimiter(spamchan, limitchan, t, style, bw)
	tchan := newTestChannels()
	return &TestFixture{
		stimuli:      stimuli,
		dampchan:     sut.Limited,
		spamchan:     sut.Spammy,
		channels:     tchan,
		config:       tcnf,
		expectations: expectations,
	}
}

func (tf *TestFixture) Test(t *testing.T) {
	var output Record = make(map[int]uint8)
	go Poll(output, tf.dampchan, tf.channels, tf.config)
	go Spam(tf.stimuli, tf.spamchan, tf.channels, tf.config)
	standby := tf.config.PulsePeriod * time.Duration(tf.config.ReservationPulses)
	<-time.After(standby)
	tf.channels.ReadySetGo <- GONOW //todo better testchan
	<-tf.channels.Finished          //hang until signalled to stop

	tf.assert(t, output)

	defer tf.close()
}

func (tf *TestFixture) assert(t *testing.T, output Record) {
	//todo display count of outputs
	//todo display count of stimuli
	//todo make assertion on pulse gaps
	//todo assert count is less than stimuli
	//todo assert new new/unknown output
	//todo assert output count is not zero

	//tf.expectations
}

func (tf *TestFixture) close() {
	close(tf.channels.ReadySetGo)
	close(tf.channels.Finished)
	//close(tf.dampchan)	// todo check closing channels https://go101.org/article/channel-closing.html
	//close(tf.spamchan)	// todo check closing channels https://go101.org/article/channel-closing.html
}
