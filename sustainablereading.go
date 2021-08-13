package sustainablereading

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/enriquebris/goconcurrentqueue"
)

type Type int

const (
	Data Type = iota
	Pause
	Error
	SysError
	Abandon
)

type Event struct {
	Kind Type
	Url  string
	Data interface{}
	Err  error
}

type Config struct {
	Timeout int
	Queue   *goconcurrentqueue.FIFO
	Msg     chan Event
	Control chan Type
}

func NewSustainableReading(timeout int, ch chan Event) *Config {
	ret := &Config{
		Timeout: timeout,
		Queue:   goconcurrentqueue.NewFIFO(),
		Msg:     ch,
		Control: make(chan Type)}
	go ret.Process()

	return ret
}

func (sr *Config) Add(url string) {
	sr.Queue.Enqueue(url)
}

func (sr *Config) Stop() {
	sr.Control <- Abandon
}

func (sr *Config) Process() {
	ch := make(chan string)
	onTimer := false
	for {
		select {
		case externalAction := <-sr.Control:
			if externalAction == Abandon {
				break
			}
		case urlToQueue := <-ch:
			sr.Queue.Enqueue(urlToQueue)

			if onTimer == false {
				sr.Msg <- Event{Kind: Pause}
				<-time.After(time.Duration(sr.Timeout) * time.Second)
			}

			onTimer = true
		default:
			onTimer = false

			if sr.Queue.GetLen() > 0 {
				url, err := sr.Queue.Dequeue()

				if err != nil {
					sr.Msg <- Event{Kind: SysError, Url: url.(string), Err: err}
					continue
				}

				go Read(url.(string), ch, sr.Msg)
			}
		}
	}
}

func Read(url string, ch chan<- string, msg chan Event) {
	resp, err := http.Get(url)

	if err != nil {
		msg <- Event{Kind: Error, Url: url, Err: err}
		ch <- url
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg <- Event{Kind: Error, Url: url, Data: resp.StatusCode}
		ch <- url
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		msg <- Event{Kind: Error, Url: url, Err: err}
		ch <- url
		return
	}

	msg <- Event{Kind: Data, Url: url, Data: body}
}
