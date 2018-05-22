package r8lmt

import (
	"time"
	"testing"
)

type TestConfig struct {
	ReservationPulses int
	PulsePeriod time.Duration
	MaxPulses int
	TimeOut time.Duration
}

const defaultTimeOut = time.Minute

const defaultPulsePeriod = time.Microsecond*50

func newTestConfig(reservationPulses int, pulsePeriod time.Duration, maxPulses int, timeOut time.Duration) *TestConfig {
	return &TestConfig{
		ReservationPulses:reservationPulses,
		MaxPulses:maxPulses,
		PulsePeriod:pulsePeriod,
		TimeOut:timeOut,
	}
}

func newTestChannels() *TestChannels {
	return &TestChannels{
		ReadySetGo:make(chan uint8),
		Finished:make(chan bool),
	}
}

func newSUT(reserveFirst bool) func(chan interface{}, *Config) chan interface{} {
	if reserveFirst {
		return NewReserveFirstSpamChan
	}
	return NewAdmitFirstSpamChan
}

type TestFixture struct {
	dampchan chan interface{}
	spamchan chan interface{}
	channels *TestChannels
	config *TestConfig
	stimuli Record
	expectations Record
}

func NewTestFixture(stimuli map[int]uint8, reservationPulses int, maxPulses int, reserveFirst bool, expandableReservation bool, repetitive bool, expectations map[int]uint8) *TestFixture {
	tcnf := newTestConfig(reservationPulses, defaultPulsePeriod, maxPulses, defaultTimeOut)
	config := &Config{
		WillAdmitAfter:repetitive,
		IsReservationExpandable:expandableReservation,
		Reservation:time.Duration(tcnf.ReservationPulses) * time.Duration(tcnf.PulsePeriod),
	}
	sut := newSUT(reserveFirst)
	dampchan := make(chan interface{})
	var spamchan chan interface{}
	spamchan = sut(dampchan, config)
	tchan := newTestChannels()
	return &TestFixture{
		stimuli:stimuli,
		dampchan:dampchan,
		spamchan:spamchan,
		channels:tchan,
		config:tcnf,
		expectations:expectations,
	}
}

func (tf *TestFixture) Test(t *testing.T) {
	var output Record = make(map[int]uint8)
	go Poll(output, tf.dampchan, tf.channels, tf.config)
	go Spam(tf.stimuli, tf.spamchan, tf.channels, tf.config)
	standby := tf.config.PulsePeriod*time.Duration(tf.config.ReservationPulses)
	<-time.After(standby)
	tf.channels.ReadySetGo<- GONOW
	<-tf.channels.Finished //hang until signalled to stop

	//todo compare output and expectations
	tf.assert(t, output)

	defer tf.close()
}

func (tf *TestFixture) assert(t *testing.T, output Record) {
	//todo compare output and tf.expectations
	//tf.expectations
}

func (tf *TestFixture) close() {
	close(tf.channels.ReadySetGo)
	close(tf.channels.Finished)
	close(tf.dampchan)
	close(tf.spamchan)
}
