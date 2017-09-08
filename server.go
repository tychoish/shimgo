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
	"strings"
	"sync"
	"time"
)

type shimServer struct {
	backend          backend
	supportedFormats []Format
	running          bool
	terminated       bool
	pid              int
	uri              string
	workingDirectory string
	errors           []string
	terminate        chan struct{}
	closed           chan struct{}
	sync.RWMutex
}

func newServer(b backend) *shimServer {
	server := &shimServer{
		backend: b,
	}
	server.setup()

	return server
}

func (s *shimServer) setup() {
	// unsafe, must be called by someone who has exclusive access
	// to the struct

	s.supportedFormats = []Format{}
	s.running = false
	s.terminated = false
	s.pid = 0
	s.uri = string(s.backend)
	s.errors = []string{}
	s.terminate = make(chan struct{})
	s.closed = make(chan struct{})

	tmpdir, err := ioutil.TempDir("", "shimgo-")
	if err != nil {
		s.errors = append(s.errors, err.Error())
	}
	s.workingDirectory = tmpdir
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
			close(ready)
			return
		}

		if err := s.backend.writeFiles(s.workingDirectory); err != nil {
			s.errors = append(s.errors, err.Error())
			s.Unlock()
			close(ready)
			return
		}

		defer os.RemoveAll(s.workingDirectory)

		cmd := s.backend.getCommand(s.workingDirectory)
		if cmd == nil {
			s.errors = append(s.errors, "unsupported backend")
			s.Unlock()
			close(ready)
			return
		}

		err := cmd.Start()

		// should replace the sleep with a better way to make
		// sure that the process is running
		time.Sleep(500 * time.Millisecond)

		if err != nil {
			s.errors = append(s.errors, err.Error())
			s.Unlock()
			close(ready)
			return
		}

		s.pid = cmd.Process.Pid

		s.running = true
		s.Unlock()

		close(ready)

		<-s.terminate
		cmd.Process.Kill()

		s.Lock()
		defer s.Unlock()
		s.terminated = true
		s.running = false
		s.pid = 0

		close(s.closed)
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

	s.terminateServer()
}

func (s *shimServer) terminateServer() {
	s.RLock()
	close(s.terminate)
	s.RUnlock()
	<-s.closed
}

func (s *shimServer) reset() {
	s.stop()

	s.Lock()
	defer s.Unlock()
	s.setup()
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

func (s *shimServer) formatIsSupported(f Format) bool {
	s.RLock()
	defer s.RUnlock()

	for _, supportedFormat := range s.supportedFormats {
		if f == supportedFormat {
			return true
		}
	}

	return false
}

func (s *shimServer) getURI(path string) string {
	s.RLock()
	defer s.RUnlock()

	return strings.Join([]string{s.uri, path}, "/")
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
	s.startIfNeeded()

	response, err := http.DefaultClient.Post(s.getURI(string(format)), "text/plain", bytes.NewReader(input))
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return nil, errors.New(response.Status)
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
	if s.formatIsSupported(format) {
		return true
	}

	s.startIfNeeded()

	response, err := http.DefaultClient.Get(s.getURI("support/" + string(format)))
	if err != nil {
		return false
	}

	if response.StatusCode != 200 {
		return false
	}

	s.Lock()
	s.supportedFormats = append(s.supportedFormats, format)
	s.Unlock()

	return true
}
