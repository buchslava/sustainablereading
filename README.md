# sustainablereading - A simple approach to painlessly collecting large amounts of information over HTTP

The package `sustainablereading` offers a public interface with methods for gathering a bunch of information over HTTP. It contains the function to retry after a failed download attempt.

## Installation

`go get github.com/buchslava/sustainablereading`

## Classes diagram

<!-- https://www.dumels.com/ -->

![Classes diagram](web/classes-diag.svg)

## How does it work

The main goal of this solution is to provide bulk HTTP reads even if an error occurs. The solution provides functionality to let you retry a failed attempt. This solution uses a [queue](https://github.com/enriquebris/goconcurrentqueue) as the main data structure. Some thread is expecting one or several new URLs to be processed and queued up. It takes an URL from the queue and tries to read after that. If an error occurs, it tries again after a while. This solution is useful in the case of the following rule. The API allows you to read information N times in period M. For example, the URL https://api.bar.foo only provides 100 successful calls in 30 minutes. Otherwise, it will throw a bad HTTP status code, say 403.

## Test Environment

In terms of functionality above, it makes sense to have a test application that allows for this behavior. So, you can use
[this one](https://github.com/buchslava/sustainablereading/blob/master/stub/stub.go)

### How to run

`cd stub`
`go run stub.go`

This solution uses port 3100 by default. If you want to use another one, do the following: "go run stub go <YOUR PORT>"

`go run stub go 3200` if you want to run the application on port 3200

In addition, you can run the application many times on different ports.



## Basic example

Here is a [basic demo example](https://github.com/buchslava/sustainablereading/blob/master/demo/demo.go)

### How to run

1. In Terminal #1

`cd stub`
`go run stub.go`

2. In Terminal #1

`cd demo`
`go run demo.go`



```go
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
```
