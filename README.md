# sustainablereading - A simple approach to painlessly collecting large amounts of information over HTTP

The package `sustainablereading` offers a public interface with methods for gathering a bunch of information over HTTP. It contains the function to retry after a failed download attempt.

## Installation

`go get github.com/buchslava/sustainablereading`

## How does it work

The main goal of this solution is to provide bulk HTTP reads even if an error occurs. The solution provides functionality to let you retry a failed attempt. This solution uses a [queue](https://github.com/enriquebris/goconcurrentqueue) as the main data structure. Some thread is expecting one or several new URLs to be processed and queued up. It takes an URL from the queue and tries to read after that. If an error occurs, it tries again after a while. This solution is useful in the case of the following rule. The API allows you to read information N times in period M. For example, the URL https://api.bar.foo only provides 100 successful calls in 30 minutes. Otherwise, it will throw a bad HTTP status code, say 403.

## Test Environment

In terms of functionality above, it makes sense to have a test application that allows for this behavior. So, you can use
[this one](https://github.com/buchslava/sustainablereading/blob/master/stub/stub.go)

### How to run

```
cd stub
go run stub.go
```

This solution uses port 3100 by default. If you want to use another one, do the following:

```
go run stub go <YOUR_PORT>
```

Run `go run stub go 3200` if you want to run the application on port 3200

Also [here](https://github.com/buchslava/sustainablereading/blob/master/stub/stub.go#L14-L15) you can change the rule regarding reading restrictions.

In addition, you can run the application many times on different ports.

### How to use

The Test Environment application works as follows. It creates 99 endpoints `http://localhost/data1` -- `http://localhost/data99`. Each of them prints a [random string and gives a status code 200](https://github.com/buchslava/sustainablereading/blob/master/stub/stub.go#L54-L55) in case of a successful call. Otherwise, it gives a [403 status code and a "forbidden"](https://github.com/buchslava/sustainablereading/blob/master/stub/stub.go#L50-L51) message. So, if you run the app and call, say, `http://localhost/data1` 4 times immediately it will be ok. The fifth time will fail. You got a proper result after some timeout. It's 10 seconds according to the demo application (see below).

## The basic example

Here is a [basic demo example](https://github.com/buchslava/sustainablereading/blob/master/demo/basic/demo.go)

One important thing. If you want to run the demo application, you must first run the Test Environment application because the demo application works [with it](https://github.com/buchslava/sustainablereading/blob/master/demo/basic/demo.go#L20).

### How to run

- In Terminal #1

```
cd stub
go run stub.go
```

- In Terminal #2

```
cd basic/demo
go run demo.go
```

### Brief explanation

Here is the complete example code.

```go
package main

import (
	"fmt"
	"time"

	"github.com/buchslava/sustainablereading"
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

Let's clarify some important points.

1. First of all, you need to create a communication channel. It's mandatory because you need to receive messages from the solution.

```go
ch := make(chan sustainablereading.Event)
```

2. Second, you have to instantiate an object that represents the main logic

```go
sr := sustainablereading.NewSustainableReading(10, ch)
```

and pass a communication channel into it.
Pay attention to the first parameter (10). This means waiting 10 seconds after trying to fall. You can of course choose another one, say 100 or 5 seconds ...

3. Now you can tell the main logic about new URL (URLs)

```go
sr.Add("http://localhost:3100/data1")
sr.Add("http://localhost:3100/data2")
//...
```

You can also add a new url a little later when the main logic works.

4. It's time to interact with the main logic

```go
	for {
		select {
		case msg := <-ch:
    // do message processing here
		default:
    // ...
		}
	}

```

5. There is the following logic in the [basic example](https://github.com/buchslava/sustainablereading/blob/master/demo/basic/demo.go)


- Make a communication channel, main logic object and add 50 URLs to be processed
- Wait for messages from the main logic: `case msg := <-ch:`

```go
// ...
const (
	Total = 50
)

// ...

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
   // ...
   default:
		if current > Total {
			sr.Stop()
			break Loop
		}
	}
}
```

- If a URL processed successfully print the result and increase a counter:

```go
if msg.Kind == sustainablereading.Data {
	fmt.Println(TimeLabel(), current, "of", Total, msg.Url, string(msg.Data.([]byte)))
	current = current + 1
}
```

- If there are no messages and all of the URLs are processed successfully then break a main loop:

```go
default:
 if current > Total {
  	sr.Stop()
	  break Loop
  }
```

## Test results

Let's get started...

- In Terminal #1

```
cd stub
go run stub.go
```

- In Terminal #2

```
cd demo/basic
go run demo.go
```

We will get the result in Terminal # 2 something like this

```
go run demo.go
16:44:37 1 of 50 http://localhost:3100/data26 rDfALKPUHONspAAQpVFj
16:44:37 2 of 50 http://localhost:3100/data25 EbLUAialtVHPwthwniJT
16:44:37 3 of 50 http://localhost:3100/data27 AJiwbSilKCsjRnRJcsmp
16:44:37 4 of 50 http://localhost:3100/data28 xhGdjtFOdnvDjKiYpPyR
16:44:37 ...
16:44:47 ...
16:44:57 ...
...................................................................
16:49:48 47 of 50 http://localhost:3100/data45 TZsHpNBAwoRAASrfYESQ
16:49:48 48 of 50 http://localhost:3100/data42 rkoTROLiQQKhpDqHcMhA
16:49:48 ...
16:49:58 ...
16:50:08 49 of 50 http://localhost:3100/data24 MaqYEojAjYwfssJDcqDm
16:50:08 50 of 50 http://localhost:3100/data47 XVrmHoFxXtjoAhyYCvaw
```

Let's analyze what we got. Please note that all HTTP calls are indeed asynchronous. They are presented as separate go routines. As you can see, the first four calls were successful. And after that, we have 3 pauses of 10 seconds. This behavior is 100% consistent with our Test Environment application. The future activities have similar behavior. The summary is as follows. We started the process at 4:44:37 PM and finished at 4:50:08 PM.

## Multi-API example

I found that the above basic example is a bit artificial because in real life we ​​have to process data from different APIs. Plus, most of them have their own unique rules. The example [below](https://github.com/buchslava/sustainablereading/blob/master/demo/multi/demo.go) illustrates this case. So, we have several APIs: `Api#1` (http://localhost:3100) and `Api#2` (http://localhost:3200).

```go
package main

import (
	"fmt"
	"time"

	"github.com/buchslava/sustainablereading"
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
```

We will see something like this if we run the solution:

```
go run demo.go
18:30:17 API#1 1 of 35 http://localhost:3100/data12 dFuAGuftnxbOpjsAUpYW
18:30:17 API#1 2 of 35 http://localhost:3100/data18 vIDYzlUJeQbNTqOPxsiZ
18:30:17 API#1 3 of 35 http://localhost:3100/data13 JKEfxdrKLySWlAUGHDIk
18:30:17 API#1 4 of 35 http://localhost:3100/data4 xkcGEnsKsFYDDlQASQGV
18:30:17 API#1 ...
18:30:17 API#2 1 of 15 http://localhost:3200/data2 QtKphHuYrCZJIVkYSNEj
18:30:17 API#2 2 of 15 http://localhost:3200/data3 nmIFSixenIzOJMjnajRu
18:30:17 API#2 3 of 15 http://localhost:3200/data4 UCfDoDRRhInESxSErfve
18:30:17 API#2 4 of 15 http://localhost:3200/data5 PVgwFnMCisLcIMjTDMkD
18:30:17 API#2 ...
18:30:27 API#1 ...
........................................................................
18:32:17 API#2 14 of 15 http://localhost:3200/data12 fYsJJQRlGqTpzvDRDNEx
18:32:17 API#2 15 of 15 http://localhost:3200/data1 kBcBwRBzXQLJGXtlnrmh
18:32:17 API#1 17 of 35 http://localhost:3100/data26 rleHzuTMhTbcqLCwKqiL
18:32:17 API#1 18 of 35 http://localhost:3100/data14 EiCGCtqJurlLizWMFpzv
........................................................................
18:34:07 API#1 ...
18:34:17 API#1 33 of 35 http://localhost:3100/data34 xbWyUoHYeopIXSMEeKAn
18:34:17 API#1 34 of 35 http://localhost:3100/data8 jpXupQKCTCsIPwgsGdKL
```

## Classes diagram

<!-- https://www.dumels.com/ -->

![Classes diagram](images/classes-diag.svg)
