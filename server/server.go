package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/garyburd/redigo/redis"
)

// Runner runs the code
type Runner struct {
	Ext    string `json:"ext"`
	Source string `json:"source"`
}

// Run the code in the container
func (r *Runner) Run(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/octet-stream")
	cmd := exec.Command("docker", "run", "-i", "koderunr", r.Ext, r.Source)

	pipeReader, pipeWriter := io.Pipe()
	defer pipeWriter.Close()
	defer pipeWriter.Close()

	cmd.Stdout = pipeWriter
	cmd.Stderr = pipeWriter

	// Doing the streaming
	go func() {
		buffer := make([]byte, 256)
		for {
			n, err := pipeReader.Read(buffer)
			if err != nil {
				if err != io.EOF {
					pipeReader.Close()
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				}
				break
			}

			data := buffer[0:n]
			w.Write(data)

			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}

			for i := 0; i < n; i++ {
				buffer[i] = 0
			}
		}
	}()

	cmd.Run()
}

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

	runner.Run(w)
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

	// conn.Do("SET", 123, 45678)
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
	http.HandleFunc("/", s.handleRunCode)
	http.HandleFunc("/register/", s.handleReg)
	http.ListenAndServe(":8080", nil)
}
