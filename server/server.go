package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/garyburd/redigo/redis"
)

// Server is the abstraction of a koderunr web api
type Server struct {
	redisPool *redis.Pool
}

func (s *Server) handleRunCode(w http.ResponseWriter, r *http.Request) {
	uuid := r.FormValue("uuid")

	conn := s.redisPool.Get()
	defer conn.Close()

	value, err := redis.Bytes(conn.Do("GET", uuid))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot GET: %v\n", err)
		http.Error(w, "Internal error.\n", 500)
		return
	}
	runner := &Runner{}
	json.Unmarshal(value, runner)

	isEvtStream := r.FormValue("evt") == "true"
	runner.Run(w, isEvtStream)
}

func (s *Server) handleReg(w http.ResponseWriter, r *http.Request) {
	runner := Runner{
		r.FormValue("ext"),
		r.FormValue("source"),
	}

	bts, _ := json.Marshal(&runner)
	strj := string(bts)

	cmd := exec.Command("uuidgen")
	output, _ := cmd.Output()
	uuid := strings.TrimSuffix(string(output), "\n")

	conn := s.redisPool.Get()
	defer conn.Close()

	_, err := conn.Do("SET", uuid, strj)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		http.Error(w, "A serious error has occured.", 500)
		return
	}

	fmt.Fprint(w, uuid)
}

func main() {
	redisPool := redis.NewPool(func() (redis.Conn, error) {
		conn, err := redis.Dial("tcp", ":6379")
		if err != nil {
			return nil, err
		}
		return conn, err
	}, 5)

	s := &Server{
		redisPool: redisPool,
	}

	http.Handle("/", http.FileServer(http.Dir("static")))
	http.HandleFunc("/run", s.handleRunCode)
	http.HandleFunc("/register/", s.handleReg)
	http.ListenAndServe(":8080", nil)
}
