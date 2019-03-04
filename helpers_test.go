package r8lmt

import (
	"fmt"
	"time"
)

var isTestingVerbose = true

var printout = func(msg string) {
	if isTestingVerbose {
		fmt.Println(msg)
	}
}

const timeStampFormat = "15:04:05.999999999"

func TimeStamp() string {
	return time.Now().Format(timeStampFormat)
}

const tickFormat = "05.999999999"

func TickStamp() string {
	return time.Now().Format(tickFormat)
}

const GONOW = uint8(2 ^ 8 - 1)

const BLANK = 0

type Record map[int]uint8

type TestChannels struct {
	ReadySetGo chan uint8
	Finished   chan bool
}

func spamStimulus(stimuli Record, key int, spammy chan<- interface{}) {
	stimulus := stimuli[key]
	if stimulus != BLANK {
		spammy <- stimulus
		str := fmt.Sprintf("in=%d  ", stimulus) + TimeStamp()
		printout(str)
	}
}

func Spam(stimuli Record, spammy chan<- interface{}, tchan *TestChannels, cnf *TestConfig) { //spam routine
	begin := time.Now()
	pulse := time.Tick(cnf.PulsePeriod)
	var txlog = spamStimulus
	var fin chan<- bool = tchan.Finished
	var readyToSpam <-chan uint8 = tchan.ReadySetGo
	time.Sleep(110) //w/o sleep tests will fail
SpamReady:
	for {
		select {
		case signal := <-readyToSpam:
			if signal == GONOW {
				printout("spam confirmed ready")
				break SpamReady
			}
		default: //nonblocking //do nothing and keep waiting
		}
	}
	printout("start spamming... " + TimeStamp())
	cnt := 0
	for {
		select {
		case <-pulse:
			txlog(stimuli, cnt, spammy)
			cnt++
		default:
			done := cnt >= cnf.MaxPulses || time.Since(begin) >= cnf.TimeOut
			if done {
				fin <- true
				printout("stop spamming @ " + TimeStamp())
				return
			}
		}
	}

}

func Poll(output Record, dampchan <-chan interface{}, tchan *TestChannels, cnf *TestConfig) {
	printout("init polling @ " + TimeStamp())
	begin := time.Now()
	var accum = make(map[int64]uint8)
	var coredump = func() {
		printout("dump @ " + time.Since(begin).String())
		for ns, rec := range accum {
			str := fmt.Sprintf("time[%d ns] : msg=%d", ns, rec)
			printout(str)
		}
	}
	var printstop = func() {
		printout("stop polling @ " + TimeStamp())
		coredump()
	}
	pulse := cnf.PulsePeriod
	dumpperiod := time.Duration(cnf.ReservationPulses) * pulse
	var donePolling <-chan bool = tchan.Finished
	var readyToPoll <-chan uint8 = tchan.ReadySetGo
	printout("are we there yet?")
PollReady:
	for {
		select {
		case signal := <-readyToPoll:
			if signal == GONOW {
				printout("poll confirmed ready")
				break PollReady
			}
		default: //nonblocking //do nothing and keep waiting
		}
	}
	printout("start polling..." + TimeStamp())
	cnt := 0
	for { //record dampchan, listen for done
		select {
		case <-time.After(pulse):
			cnt++
		case <-time.After(dumpperiod):
			go coredump()
		case msg, ok := <-dampchan:
			if !ok {
				printstop()
				return
			}
			switch v := msg.(type) {
			case uint8: //record to output
				if v != BLANK {
					str := fmt.Sprintf("@cnt[%d] := %d  ", cnt, v) + TimeStamp()
					printout(str)
					accum[time.Now().UnixNano()] = v
					output[cnt] = v
				}
			}
		case fin := <-donePolling:
			if fin {
				printstop()
				return
			}
		default:
			done := cnt >= cnf.MaxPulses || time.Since(begin) >= cnf.TimeOut
			if done {
				printstop()
				return
			}
		}
	}
}
