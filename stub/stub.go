package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

const (
	QTY_LIMIT  = 4
	TIME_LIMIT = 30
)

type App struct {
	Router *mux.Router
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

func Allow() bool {
	now := time.Now()
	if len(times) <= 1 {
		return true
	}

	qty := 0
	for i := len(times) - 1; i >= 0; i-- {
		if now.Sub(times[i]).Seconds() < TIME_LIMIT {
			qty = qty + 1
		}
	}
	return qty < QTY_LIMIT
}

func (a *App) getRandom(w http.ResponseWriter, r *http.Request) {
	if Allow() == false {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "forbidden")
	} else {
		times = append(times, time.Now())
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, RandStringRunes(20))
	}
}

func (a *App) Initialize() {
	rand.Seed(time.Now().UnixNano())

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
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	a.Initialize()
	a.Run(port)
}
