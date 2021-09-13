package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type Type int

const (
	NoRetryAfter Type = iota
	RetryAfterSeconds
	RetryAfterHttpTime
)

const (
	QTY_LIMIT  = 4
	TIME_LIMIT = 30
)

type App struct {
	Router           *mux.Router
	RequestDelay     int
	RequestDelayType Type
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var times = []time.Time{}

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func Allow() (bool, int) {
	now := time.Now()
	if len(times) <= 1 {
		return true, 0
	}

	qty := 0
	rest := 0.0
	for i := len(times) - 1; i >= 0; i-- {
		if now.Sub(times[i]).Seconds() < TIME_LIMIT {
			rest = now.Sub(times[i]).Seconds()
			qty = qty + 1
		}
	}

	return qty < QTY_LIMIT, int(rest)
}

func (a *App) getRandom(w http.ResponseWriter, r *http.Request) {
	if a.RequestDelay > 0 {
		time.Sleep(time.Duration(a.RequestDelay) * time.Second)
	}

	isAllowed, rest := Allow()
	if rest > 0 && a.RequestDelayType == RetryAfterSeconds {
		w.Header().Set("Retry-After", strconv.Itoa(rest))
	}
	if rest > 0 && a.RequestDelayType == RetryAfterHttpTime {
		w.Header().Set("Retry-After", time.Now().Add(time.Duration(rest)*time.Second).UTC().Format(http.TimeFormat))
	}
	if isAllowed == false {
		// put the header
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "forbidden")
	} else {
		times = append(times, time.Now())
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, RandStringRunes(20))
	}
}

func (a *App) Initialize(delay int, retryAfter Type) {
	rand.Seed(time.Now().UnixNano())

	a.RequestDelay = delay
	a.RequestDelayType = retryAfter
	a.Router = mux.NewRouter()
	a.initializeRouters()
	http.Handle("/", a.Router)
}

func (a *App) initializeRouters() {
	for i := 1; i < 100; i++ {
		a.Router.HandleFunc(fmt.Sprintf("/data%d", i), a.getRandom).Methods("GET")
	}
}

func (a *App) Run(port string) {
	fmt.Println(fmt.Sprintf("Started on port %s", port))
	fmt.Println(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func main() {
	a := App{}

	port := "3100"
	delay := 0
	retryAfter := NoRetryAfter
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	if len(os.Args) > 2 {
		port = os.Args[1]
		i, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
		delay = i
		if len(os.Args) > 3 {
			if os.Args[3] == "SEC" {
				retryAfter = RetryAfterSeconds
			}
			if os.Args[3] == "HTTP_DATE" {
				retryAfter = RetryAfterHttpTime
			}
		}
	}

	a.Initialize(delay, retryAfter)
	a.Run(port)
}
