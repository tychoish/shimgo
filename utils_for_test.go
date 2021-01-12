package shimgo

import (
	"os"
	"testing"
)

func assert(t *testing.T, condition bool, args ...interface{}) {
	if condition {
		return
	}

	t.Error(args...)
}

func require(t *testing.T, condition bool, args ...interface{}) {
	if condition {
		return
	}

	t.Fatal(args...)
}

func cleanup(t *testing.T, s *shimServer) {
	// this is basically a re-implementation of s.stop() but with assertions if thins go wrong.

	if s.workingDirectory != "" {
		err := os.RemoveAll(s.workingDirectory)
		require(t, err == nil, "cleanup working directory", err)
	}

	if s.pid != 0 {
		proc, err := os.FindProcess(s.pid)
		if err != nil {
			err = proc.Kill()
			require(t, err == nil, "kill service proc", err)
		}
	}
	s.stop()
	s.Lock()
	defer s.Unlock()
	s.setup()
}
