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
		http.Error(w, "The source code doesn't exist", 422)
		return
	}

	// Started running code
	runner := &Runner{}
	json.Unmarshal(value, runner)

	isEvtStream := r.FormValue("evt") == "true"
	client := NewClient(runner, s.redisPool.Get(), uuid)

	go client.Write(w, isEvtStream)
	client.Run()

	// Purge the source code
	_, err = conn.Do("DEL", uuid)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to purge the source code for %s - %v\n", uuid, err)
	}
}

func (s *Server) handleReg(w http.ResponseWriter, r *http.Request) {
	runner := Runner{
		r.FormValue("ext"),
		r.FormValue("source"),
		r.FormValue("version"),
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

func (s *Server) handleStdin(w http.ResponseWriter, r *http.Request) {
	input := r.FormValue("input")
	uuid := r.FormValue("uuid")

	conn := s.redisPool.Get()
	defer conn.Close()

	conn.Do("PUBLISH", uuid+"#stdin", input)

	fmt.Fprintf(w, "")
}
