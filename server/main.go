package main

import (
	"flag"
	"net/http"

	"github.com/garyburd/redigo/redis"
)

var servingStatic bool
var runnerThrottleNum int

func init() {
	flag.BoolVar(&servingStatic, "static", false, "if using Go server hosting static files")
	flag.IntVar(&runnerThrottleNum, "runner_throttle", 4, "Limit the max throttle for the runners")
	flag.Parse()
}

func main() {
	Runnerthrottle = make(chan struct{}, runnerThrottleNum)

	redisPool := redis.NewPool(func() (redis.Conn, error) {
		conn, err := redis.Dial("tcp", ":6379")
		if err != nil {
			return nil, err
		}
		return conn, err
	}, 4)

	s := &Server{
		redisPool: redisPool,
	}

	if servingStatic {
		http.Handle("/", http.FileServer(http.Dir("static")))
	}

	http.HandleFunc("/langs/", s.HandleLangs)
	http.HandleFunc("/run/", s.HandleRunCode)
	http.HandleFunc("/save/", s.HandleSaveCode)
	http.HandleFunc("/register/", s.HandleReg)
	http.HandleFunc("/stdin/", s.HandleStdin)
	http.HandleFunc("/fetch/", s.HandleFetchCode)
	http.ListenAndServe(":8080", nil)
}
