package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"log/syslog"

	"github.com/Sirupsen/logrus"
	logrus_syslog "github.com/Sirupsen/logrus/hooks/syslog"
	"github.com/garyburd/redigo/redis"
)

// Server is the abstraction of a koderunr web api
type Server struct {
	redisPool     *redis.Pool
	logger        *logrus.Logger
	servingStatic bool
}

// NewServer create a new Server struct
func NewServer(maxRedisConn int, servingStatic bool) *Server {
	redisPool := redis.NewPool(func() (redis.Conn, error) {
		conn, err := redis.Dial("tcp", ":6379")
		if err != nil {
			return nil, err
		}
		return conn, err
	}, maxRedisConn)

	log := logrus.New()
	hook, err := logrus_syslog.NewSyslogHook("", "", syslog.LOG_INFO, "[KodeRunr Service]")

	if err != nil {
		panic(err)
	}
	log.Hooks.Add(hook)

	return &Server{
		redisPool:     redisPool,
		logger:        log,
		servingStatic: servingStatic,
	}
}

// Serve start serving http requests
func (s *Server) Serve(scope string, port int) {
	s.logger.Infof("KodeRunr starting on port: %d", port)

	if s.servingStatic {
		http.Handle("/", http.FileServer(http.Dir("static")))
	}

	http.HandleFunc(scope+"langs/", s.HandleLangs)
	http.HandleFunc(scope+"run/", s.HandleRunCode)
	http.HandleFunc(scope+"save/", s.HandleSaveCode)
	http.HandleFunc(scope+"register/", s.HandleReg)
	http.HandleFunc(scope+"stdin/", s.HandleStdin)
	http.HandleFunc(scope+"fetch/", s.HandleFetchCode)

	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

// HandleRunCode streams the running program output to the frontend
func (s *Server) HandleRunCode(w http.ResponseWriter, r *http.Request) {
	uuid := r.FormValue("uuid")

	conn := s.redisPool.Get()
	defer conn.Close()

	// Fetch the code into runner from Redis
	runner, err := FetchCode(uuid, conn)
	if err != nil {
		s.logger.Infof("Source code cannot be found in redis - %v", err)
		http.Error(w, "Cannot find the source code for some reason", 422)
		return
	}

	// for close the container right away after the request is halted
	closeNotifier := w.(http.CloseNotifier).CloseNotify()
	runner.closeNotifier = closeNotifier
	runner.logger = s.logger

	isEvtStream := r.FormValue("evt") == "true"
	client := NewClient(runner, s.redisPool.Get(), uuid)

	go client.Write(w, isEvtStream)
	client.Run()

	// Purge the source code
	_, err = conn.Do("DEL", uuid+"#run")
	if err != nil {
		s.logger.Errorf("Failed to purge the source code for %s - %v", uuid, err)
	}
}

// HandleSaveCode saves the source code and returns a ID.
func (s *Server) HandleSaveCode(w http.ResponseWriter, r *http.Request) {
	runner := Runner{
		Lang:    r.FormValue("lang"),
		Source:  r.FormValue("source"),
		Version: r.FormValue("version"),
	}

	bts, _ := json.Marshal(&runner)
	strj := string(bts)

	codeID := r.FormValue("codeID")
	if codeID == "" {
		codeID = NewRandID(10)
	}

	conn := s.redisPool.Get()
	defer conn.Close()

	_, err := conn.Do("SET", codeID+"#snippet", strj)
	if err != nil {
		s.logger.Errorf("Failed to store code snippet: %v", err)
		http.Error(w, "A serious error has occured.", 500)
		return
	}

	fmt.Fprintf(w, codeID)
}

// HandleFetchCode loads the code by codeID and returns the source code to user
// Only used by web interface at the moment.
func (s *Server) HandleFetchCode(w http.ResponseWriter, r *http.Request) {
	codeID := r.FormValue("codeID")

	conn := s.redisPool.Get()
	defer conn.Close()

	value, err := redis.Bytes(conn.Do("GET", codeID+"#snippet"))
	if err != nil {
		s.logger.Errorf("Cannot get code snippet: %v", err)
		http.Error(w, "The source code doesn't exist", 422)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write(value)
}

// HandleReg fetch the code from the client and save it in Redis
func (s *Server) HandleReg(w http.ResponseWriter, r *http.Request) {
	runner := Runner{
		Lang:    r.FormValue("lang"),
		Source:  r.FormValue("source"),
		Version: r.FormValue("version"),
		Timeout: 15,
	}

	bts, _ := json.Marshal(&runner)
	strj := string(bts)

	cmd := exec.Command("uuidgen")
	output, _ := cmd.Output()
	uuid := strings.TrimSuffix(string(output), "\n")

	conn := s.redisPool.Get()
	defer conn.Close()

	_, err := conn.Do("SET", uuid+"#run", strj)
	if err != nil {
		s.logger.Errorf("Cannot register the code: %v", err)
		http.Error(w, "A serious error has occured.", 500)
		return
	}

	fmt.Fprint(w, uuid)
}

// HandleStdin consumes the stdin from the client side
func (s *Server) HandleStdin(w http.ResponseWriter, r *http.Request) {
	input := r.FormValue("input")
	uuid := r.FormValue("uuid")

	conn := s.redisPool.Get()
	defer conn.Close()

	conn.Do("PUBLISH", uuid+"#stdin", input)

	fmt.Fprintf(w, "")
}

//HandleLangs deals with the request for show available programming languages
func (s *Server) HandleLangs(w http.ResponseWriter, r *http.Request) {
	text := `
Supported Languages:
  Ruby - 2.3.0
  Ruby - 1.9.3-p550
  Python - 2.7.6
  Python - 3.5.0
  Swift - 2.2
  C - GCC 4.9
  Go - 1.6
  Elixir - 1.2.3
`
	text = strings.TrimSpace(text)

	fmt.Fprintf(w, "%s\n", text)
}
