package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/enriquebris/goconcurrentqueue"
)

var queue = goconcurrentqueue.NewFIFO()

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

	currentTime := time.Now()
	fmt.Println(currentTime.Format("15:04:05"), "Finished ", url, string(body))
}

func fill(ch chan<- string) {
	for i := 1; i < 21; i++ {
		ch <- fmt.Sprintf("http://localhost:3100/data%d", i)
	}
}

func process() {
	ch := make(chan string)
	for {
		select {
		case badUrl := <-ch:
			queue.Enqueue(badUrl)
			// fmt.Println("pause...")
			<-time.After(10 * time.Second)
		default:
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
