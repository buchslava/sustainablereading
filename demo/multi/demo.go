package main

import (
	"fmt"
	"time"

	"sustainablereading"
)

const (
	Api1Total = 35
	Api2Total = 15
)

func main() {
	chApi1 := make(chan sustainablereading.Event)
	chApi2 := make(chan sustainablereading.Event)
	srApi1 := sustainablereading.NewSustainableReading(10, chApi1)
	srApi2 := sustainablereading.NewSustainableReading(20, chApi2)
	currentApi1 := 1
	currentApi2 := 1

	for i := 1; i < Api1Total+1; i++ {
		srApi1.Add(fmt.Sprintf("http://localhost:3100/data%d", i))
	}
	for i := 1; i < Api2Total+1; i++ {
		srApi2.Add(fmt.Sprintf("http://localhost:3200/data%d", i))
	}

Loop:
	for {
		select {
		case msg := <-chApi1:
			GotMessage("1", msg, &currentApi1, Api1Total)
		case msg := <-chApi2:
			GotMessage("2", msg, &currentApi2, Api2Total)
		default:
			if currentApi1+currentApi2 > Api1Total+Api2Total {
				srApi1.Stop()
				srApi2.Stop()
				break Loop
			}
		}
	}
}

func GotMessage(api string, msg sustainablereading.Event, current *int, total int) {
	apiLabel := fmt.Sprintf("API#%s", api)

	if msg.Kind == sustainablereading.Data {
		fmt.Println(TimeLabel(), apiLabel, *current, "of", total, msg.Url, string(msg.Data.([]byte)))
		*current = *current + 1
	}
	if msg.Kind == sustainablereading.Pause {
		fmt.Println(TimeLabel(), apiLabel, "...")
	}
	if msg.Kind == sustainablereading.SysError {
		fmt.Println(TimeLabel(), apiLabel, msg.Err)
	}
}

func TimeLabel() string {
	currentTime := time.Now()
	return currentTime.Format("15:04:05")
}
