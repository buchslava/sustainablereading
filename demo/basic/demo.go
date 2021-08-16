package main

import (
	"fmt"
	"time"

	"sustainablereading"
)

const (
	Total = 50
)

func main() {
	ch := make(chan sustainablereading.Event)
	sr := sustainablereading.NewSustainableReading(10, ch)
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

func TimeLabel() string {
	currentTime := time.Now()
	return currentTime.Format("15:04:05")
}
