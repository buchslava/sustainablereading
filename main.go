package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/enriquebris/goconcurrentqueue"
)

const (
	Total = 50
)

var ops uint64
var queue = goconcurrentqueue.NewFIFO()

func TimeLabel() string {
	currentTime := time.Now()
	return currentTime.Format("15:04:05")
}

func Read(url string, ch chan<- string) {
	// fmt.Println("Trying to read ", url)

	resp, err := http.Get(url)

	if err != nil {
		// fmt.Println("HTTP error", url, err)
		ch <- url
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// fmt.Println("HTTP error", url, " status code", resp.StatusCode)
		ch <- url
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// fmt.Println("Data error", url, err)
		ch <- url
		return
	}

	atomic.AddUint64(&ops, 1)
	fmt.Println(TimeLabel(), "Finished ", url, string(body), ops, " of ", Total)
}

func Fill(ch chan<- string) {
	for i := 1; i < Total+1; i++ {
		ch <- fmt.Sprintf("http://localhost:3100/data%d", i)
	}
}

func Process() {
	ch := make(chan string)
	onTimer := false
	for {
		select {
		case badUrl := <-ch:
			queue.Enqueue(badUrl)
			if onTimer == false {
				fmt.Println(TimeLabel(), "pause...")
				<-time.After(10 * time.Second)
			}
			onTimer = true
		default:
			onTimer = false
			if queue.GetLen() > 0 {
				url, err := queue.Dequeue()

				if err != nil {
					fmt.Println(err)
					continue
				}

				go Read(url.(string), ch)
			}
		}
	}
}

func main() {
	ch := make(chan string)

	go Fill(ch)
	go Process()

	for {
		select {
		case url := <-ch:
			queue.Enqueue(url)
		default:
		}
	}
}
