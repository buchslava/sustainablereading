package sustainablereading

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync/atomic"
	"time"
	"strconv"

	. "github.com/enriquebris/goconcurrentqueue"
)

type Type int

type Readable func(url string, cb ReadCallback)

type ReadCallback func(Err error, Data interface{}, retrySeconds int)

const (
	Data Type = iota
	Pause
	Error
	SysError
	Abandon
)

type Event struct {
	Kind       Type
	Url        string
	Data       interface{}
	Err        error
	RetryAfter int
}

type Config struct {
	Timeout       int
	Queue         *FIFO
	Msg           chan Event
	Control       chan Type
	CustomReader  Readable
	Limit         int
	NumInProgress int32
}

func NewSustainableReading(timeout int, ch chan Event) *Config {
	ret := &Config{
		Timeout:       timeout,
		Queue:         NewFIFO(),
		Msg:           ch,
		Control:       make(chan Type),
		Limit:         0,
		NumInProgress: 0}
	go ret.Process()

	return ret
}

func (sr *Config) SetCustomReader(r Readable) {
	sr.CustomReader = r
}

func (sr *Config) SetLimit(l int) {
	sr.Limit = l
}

func (sr *Config) GetProcessesQuantity() int32 {
	return atomic.LoadInt32(&sr.NumInProgress)
}

func (sr *Config) IsWorking() bool {
	processes := atomic.LoadInt32(&sr.NumInProgress)

	return processes > 0 && sr.Queue.GetLen() > 0
}

func (sr *Config) IsFree() bool {
	processes := atomic.LoadInt32(&sr.NumInProgress)

	return processes <= 0 && sr.Queue.GetLen() <= 0
}

func (sr *Config) Add(url string) {
	sr.Queue.Enqueue(url)
}

func (sr *Config) Stop() {
	sr.Control <- Abandon
}

func (sr *Config) Process() {
	ch := make(chan struct {string; int})
	onTimer := false
	for {
		select {
		case externalAction := <-sr.Control:
			if externalAction == Abandon {
				break
			}
		case backToQueue := <-ch:
			sr.Queue.Enqueue(backToQueue.string)

			if onTimer == false {
				if backToQueue.int > 0 {
					sr.Msg <- Event{Kind: Pause, RetryAfter: backToQueue.int}
					<-time.After(time.Duration(backToQueue.int) * time.Second)
				} else {
					sr.Msg <- Event{Kind: Pause}
					<-time.After(time.Duration(sr.Timeout) * time.Second)
				}
			}

			onTimer = true
		default:
			onTimer = false

			NumInProgressValue := atomic.LoadInt32(&sr.NumInProgress)

			if sr.Queue.GetLen() > 0 && (sr.Limit == 0 || NumInProgressValue < int32(sr.Limit)) {
				url, err := sr.Queue.Dequeue()

				if err != nil {
					sr.Msg <- Event{Kind: SysError, Url: url.(string), Err: err}
					continue
				}

				atomic.AddInt32(&sr.NumInProgress, 1)

				if sr.CustomReader != nil {
					go sr.CustomReader(url.(string), func(Err error, Body interface{}, RetryAfter int) {
						if Err != nil {
							sr.Msg <- Event{Kind: Error, Url: url.(string), Err: err, RetryAfter: RetryAfter}
							ch <- struct {string; int}{url.(string), RetryAfter}
						} else {
							sr.Msg <- Event{Kind: Data, Url: url.(string), Data: Body}
						}
						atomic.AddInt32(&sr.NumInProgress, -1)
					})
				} else {
					go Read(url.(string), func(Err error, Body interface{}, RetryAfter int) {
						if Err != nil {
							sr.Msg <- Event{Kind: Error, Url: url.(string), Err: Err, RetryAfter: RetryAfter}
							ch <- struct {string; int}{url.(string), RetryAfter}
						} else {
							sr.Msg <- Event{Kind: Data, Url: url.(string), Data: Body}
						}
						atomic.AddInt32(&sr.NumInProgress, -1)
					})
				}
			}
		}
	}
}

func Read(url string, cb ReadCallback) {
	resp, err := http.Get(url)

	if err != nil {
		cb(err, nil, 0)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		retryTime := resp.Header.Get("Retry-After")

		if retryTime != "" {
			timeSec, timeSecErr := strconv.Atoi(retryTime)

			if timeSecErr != nil {
				timeTime, timeTimeErr := time.Parse(http.TimeFormat, retryTime)

				if timeTimeErr == nil {
					cb(errors.New(fmt.Sprintf("Wrong status code: %d", resp.StatusCode)), nil, int(timeTime.Sub(time.Now()).Seconds()))
					return
				}
	  	} else {
				cb(errors.New(fmt.Sprintf("Wrong status code: %d", resp.StatusCode)), nil, timeSec)
				return
			}
		}

		cb(errors.New(fmt.Sprintf("Wrong status code: %d", resp.StatusCode)), nil, 0)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		cb(err, nil, 0)
		return
	}

	cb(nil, body, 0)
}
