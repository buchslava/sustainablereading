package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/enriquebris/goconcurrentqueue"
)

var queue = goconcurrentqueue.NewFIFO()
var total = 50
var current = 0

func read(url string, ch chan<- string) {
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

	current = current + 1
	currentTime := time.Now()
	fmt.Println(currentTime.Format("15:04:05"), "Finished ", url, string(body), current, " of ", total)
}

func fill(ch chan<- string) {
	for i := 1; i < total+1; i++ {
		ch <- fmt.Sprintf("http://localhost:3100/data%d", i)
	}
}

func process() {
	ch := make(chan string)
	onTimer := false
	for {
		select {
		case badUrl := <-ch:
			queue.Enqueue(badUrl)
			currentTime := time.Now()
			if onTimer == false {
				fmt.Println(currentTime.Format("15:04:05"), "pause...")
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

				go read(url.(string), ch)
			}
		}
	}
}

func main() {
	ch := make(chan string)

	go fill(ch)
	go process()

	for {
		select {
		case url := <-ch:
			queue.Enqueue(url)
		default:
		}
	}
}
