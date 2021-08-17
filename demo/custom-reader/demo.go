package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	. "sustainablereading"
)

const (
	Total = 50
)

func main() {
	ch := make(chan Event)
	sr := NewSustainableReading(10, ch)
	sr.SetCustomReader(CustomReader("some additionals"))
	current := 1

	for i := 1; i < Total+1; i++ {
		sr.Add(fmt.Sprintf("http://localhost:3100/data%d", i))
	}

Loop:
	for {
		select {
		case msg := <-ch:
			if msg.Kind == Data {
				fmt.Println(TimeLabel(), current, "of", Total, msg.Url, string(msg.Data.([]byte)))
				current = current + 1
			}
			if msg.Kind == Pause {
				fmt.Println(TimeLabel(), "...")
			}
			if msg.Kind == SysError {
				fmt.Println(TimeLabel(), msg.Err)
			}
		default:
			if current > Total {
				sr.Stop()
				break Loop
			}
		}
	}
}

func CustomReader(additional interface{}) Readable {
	return func(url string, ch chan<- string, msg chan Event) {
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

		fmt.Println(fmt.Sprintf("Use CustomReader with %s", additional))
		msg <- Event{Kind: Data, Url: url, Data: body}
	}
}

func TimeLabel() string {
	currentTime := time.Now()
	return currentTime.Format("15:04:05")
}
