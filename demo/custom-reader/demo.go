package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	"strconv"

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
	return func(url string, cb ReadCallback) {
		resp, err := http.Get(url)

		if err != nil {
			cb(err, nil, 0)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			retryTime := resp.Header.Get("Retry-After")

			if retryTime != "" {
				timeSec, err1 := strconv.Atoi(retryTime)

				if err1 != nil {
					timeTime, err2 := time.Parse(http.TimeFormat, retryTime)

					if err2 == nil {
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

		fmt.Println(fmt.Sprintf("Run CustomReader with %s", additional))
		cb(nil, body, 0)
	}
}

func TimeLabel() string {
	currentTime := time.Now()
	return currentTime.Format("15:04:05")
}
