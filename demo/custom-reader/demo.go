package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"sustainablereading"
)

const (
	Total = 50
)

func main() {
	ch := make(chan sustainablereading.Event)
	sr := sustainablereading.NewSustainableReading(10, ch)
	sr.SetCustomReader(CustomReader)
	current := 1

	for i := 1; i < Total+1; i++ {
		sr.Add(fmt.Sprintf("http://localhost:3100/data%d", i))
	}

Loop:
	for {
		select {
		case msg := <-ch:
			if msg.Kind == sustainablereading.Data {
				fmt.Println(TimeLabel(), current, "of", Total, msg.Url, string(msg.Data.([]byte)))
				current = current + 1
			}
			if msg.Kind == sustainablereading.Pause {
				fmt.Println(TimeLabel(), "...")
			}
			if msg.Kind == sustainablereading.SysError {
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

func CustomReader(url string, ch chan<- string, msg chan sustainablereading.Event) {
	resp, err := http.Get(url)

	if err != nil {
		msg <- sustainablereading.Event{Kind: sustainablereading.Error, Url: url, Err: err}
		ch <- url
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg <- sustainablereading.Event{Kind: sustainablereading.Error, Url: url, Data: resp.StatusCode}
		ch <- url
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		msg <- sustainablereading.Event{Kind: sustainablereading.Error, Url: url, Err: err}
		ch <- url
		return
	}

	fmt.Println("Use CustomReader!")
	msg <- sustainablereading.Event{Kind: sustainablereading.Data, Url: url, Data: body}
}

func TimeLabel() string {
	currentTime := time.Now()
	return currentTime.Format("15:04:05")
}
