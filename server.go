package shimgo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var server *shimServer

func init() {
	server = newServer()
}

type shimServer struct {
	running          bool
	terminated       bool
	pid              int
	uri              string
	workingDirectory string
	errors           []string
	done             chan struct{}
	closed           chan struct{}
	sync.RWMutex
}

func newServer() *shimServer {
	server = &shimServer{
		done:   make(chan struct{}),
		closed: make(chan struct{}),
		uri:    "http://localhost:1414/",
	}

	tmpdir, err := ioutil.TempDir("", "shimgo-")
	server.workingDirectory = tmpdir
	server.addError(err)

	return server
}

func (s *shimServer) addError(err error) {
	if err != nil {
		s.Lock()
		defer s.Unlock()

		s.errors = append(s.errors, fmt.Sprintf("%+v", err))
	}
}

func (s *shimServer) start() {
	ready := make(chan struct{})
	go func() {
		s.Lock()
		if s.running {
			s.Unlock()
			ready <- struct{}{}
			return
		}

		if err := writePythonFiles(s.workingDirectory); err != nil {
			s.errors = append(s.errors, err.Error())
			s.Unlock()
			ready <- struct{}{}
			return
		}

		defer os.RemoveAll(s.workingDirectory)

		cmd := exec.Command(getPython2(), filepath.Join(s.workingDirectory, service))
		err := cmd.Start()

		s.pid = cmd.Process.Pid
		// should replace the sleep with a better way to make
		// sure that the process is running
		time.Sleep(500 * time.Millisecond)

		if err != nil {
			s.errors = append(s.errors, err.Error())
			s.Unlock()
			ready <- struct{}{}
			return
		}

		s.running = true
		s.Unlock()

		ready <- struct{}{}

		<-s.done
		cmd.Process.Kill()

		s.Lock()
		s.terminated = true
		s.running = false
		s.pid = 0
		s.Unlock()

		s.closed <- struct{}{}
	}()

	<-ready
}

func (s *shimServer) stop() {
	if s.hasTerminated() {
		return
	}

	if !s.isRunning() {
		return
	}

	s.done <- struct{}{}
	<-s.closed
}

func (s *shimServer) startIfNeeded() error {
	if s.isRunning() {
		return nil
	}

	if s.hasTerminated() {
		return nil
	}

	if s.hasError() {
		return s.getError()
	}

	s.start()

	return s.getError()
}

func (s *shimServer) isRunning() bool {
	s.RLock()
	defer s.RUnlock()

	return s.running
}

func (s *shimServer) hasTerminated() bool {
	s.RLock()
	defer s.RUnlock()

	return s.terminated
}

func (s *shimServer) hasError() bool {
	s.RLock()
	defer s.RUnlock()

	return len(s.errors) > 0
}

func (s *shimServer) getError() error {
	s.RLock()
	defer s.RUnlock()

	if len(s.errors) == 0 {
		return nil
	}

	return errors.New(strings.Join(s.errors, "\n"))
}

func (s *shimServer) doConversion(format Format, input []byte) ([]byte, error) {
	if format != ASCIIDOC && format != RST {
		return nil, fmt.Errorf("%s is not a supported format", format)
	}

	response, err := http.DefaultClient.Post(s.uri+string(format), "text/plain", bytes.NewReader(input))
	if err != nil {
		return nil, err
	}
	output := bytes.NewBuffer([]byte{})
	_, err = io.Copy(output, response.Body)
	response.Body.Close()
	if err != nil {
		return nil, err
	}

	data := &struct {
		Content string
		Info    string
	}{}

	if err := json.Unmarshal(output.Bytes(), data); err != nil {
		return nil, err
	}

	out := []byte(data.Content)

	if data.Info != "" {
		return out, errors.New(data.Info)
	}

	return []byte(out), nil
}

func (s *shimServer) supportsConversion(format Format) bool {
	response, err := http.DefaultClient.Get(s.uri + "support/" + string(format))
	if err != nil {
		return false
	}

	if response.StatusCode != 200 {
		return false
	}

	return true
}
