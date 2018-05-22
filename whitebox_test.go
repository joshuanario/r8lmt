package r8lmt

import (
	"testing"
)
// todo generate expectations

const defaultMaxPulses = 80

const defaultReservationPulses = 5

func TestThrottlerShortDelay(t *testing.T) {
	stimuli := map[int]uint8{
		0: 255,
		2: 254,
		3: 253,
		13: 252,
		16: 251,
		18: 250,
		22: 249,
		23: 248,
		24: 247,
	}
	expectations := stimuli //todo
	testRateLimiter(t, stimuli, 1, defaultMaxPulses, false, true, false, expectations)
}

func TestThrottlerLongDelay(t *testing.T) {
	stimuli := map[int]uint8{
		0: 255,
		2: 254,
		3: 253,
		13: 252,
		16: 251,
		18: 250,
		22: 249,
		23: 248,
		24: 247,
	}
	expectations := stimuli //todo
	testRateLimiter(t, stimuli, defaultReservationPulses, defaultMaxPulses, false, true, false, expectations)
}

func TestDebouncerFeedback(t *testing.T) {
	stimuli := map[int]uint8{
		0: 255,
		2: 254,
		3: 253,
		13: 252,
		16: 251,
	}
	expectations := stimuli //todo
	testRateLimiter(t, stimuli, defaultReservationPulses, defaultMaxPulses, false, true, false, expectations)
}

func TestDebouncerNoFeedback(t *testing.T) {
	stimuli := map[int]uint8{
		0: 255,
		2: 254,
		3: 253,
		13: 252,
		16: 251,
	}
	expectations := stimuli //todo
	testRateLimiter(t, stimuli, defaultReservationPulses, defaultMaxPulses, false, true, false, expectations)
}

func testRateLimiter(t *testing.T, stimuli map[int]uint8, reservationPulses int, maxPulses int, reserveFirst bool, expandableReservation bool, repetitive bool, expectations map[int]uint8) {
	tf := NewTestFixture(stimuli,reservationPulses, maxPulses, reserveFirst, expandableReservation, repetitive, expectations)
	tf.Test(t)
}