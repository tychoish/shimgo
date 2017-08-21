package shimgo

import (
	"sync"
)

var serverCache *servers

type servers struct {
	backends map[Format]*shimServer
	mu       sync.RWMutex
}

func init() {
	pyserver := newServer(pythonServer)
	rbserver := newServer(rubyServer)

	serverCache = &servers{
		backends: map[Format]*shimServer{
			RST:         pyserver,
			ASCIIDOC:    pyserver,
			ASCIIDOCTOR: rbserver,
		},
	}
}

func (s *servers) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, server := range s.backends {
		server.stop()
	}
}

func (s *servers) reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, server := range s.backends {
		wasRunning := server.isRunning()

		server.reset()

		if wasRunning {
			server.start()
		}
	}
}

func (s *servers) getServer(f Format) (*shimServer, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	server, ok := s.backends[f]

	if !ok {
		return nil, false
	}

	if !server.supportsConversion(f) {
		return nil, false
	}

	return server, true
}
