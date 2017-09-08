package shimgo

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
)

func TestServersHaveCorrectInitialValues(t *testing.T) {
	assert(t, serverCache != nil, "global instance is initialized")

	assert(t, len(serverCache.backends) == 3, "servers are less than expected")
	for _, server := range serverCache.backends {
		assert(t, len(server.errors) == 0, "there are no errors")
		assert(t, len(server.workingDirectory) != 0, "working is defined", server.workingDirectory)
		_, err := os.Stat(server.workingDirectory)
		assert(t, !os.IsNotExist(err), "working directory exists:", server.workingDirectory)

		assert(t, server.pid == 0, "pid is zeroed")

		cleanup(t, server)
	}
}

func TestAddError(t *testing.T) {
	for _, s := range serverCache.backends {
		assert(t, len(s.errors) == 0, "there should be no errors and there are:", len(s.errors))
		assert(t, !s.hasError(), "has error should not report error but does")
		assert(t, s.getError() == nil, "get error should be nil")

		s.addError(nil)
		assert(t, len(s.errors) == 0, "there should be no errors and there are:", len(s.errors))
		assert(t, !s.hasError(), "has error should not report error but does")
		assert(t, s.getError() == nil, "get error should be nil")

		s.addError(errors.New("foo"))
		assert(t, len(s.errors) == 1, "there should be one error and there are:", len(s.errors))
		assert(t, s.hasError(), "has error should report error but does not")
		assert(t, s.getError() != nil, "get error should not be nil")

		wg := &sync.WaitGroup{}
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for n := 0; n < 10; n++ {
					s.addError(errors.New("bar"))
				}
			}()
		}
		wg.Wait()

		assert(t, len(s.errors) == 201, "there are many errors (201), and are actually:", len(s.errors))
		assert(t, s.hasError(), "has error should report error but does not")
		assert(t, s.getError() != nil, "get error should not be nil")

		cleanup(t, s)
	}
}

func TestStartingService(t *testing.T) {
	for _, s := range serverCache.backends {
		require(t, !s.running, "server shouldn't be running at start, but is", fmt.Sprintf("%+v", s))
		assert(t, !s.isRunning(), "isRunning method should reflect that server is not yet running")
		s.start()
		require(t, s.running, "server should be running after starting, but isn't", fmt.Sprintf("%+v", s))
		assert(t, s.isRunning(), "isRunning method should reflect that server is running")
		assert(t, !s.hasTerminated(), "server should not report terminated before it has")

		cleanup(t, s)
	}
}

func TestStartIfNeeded(t *testing.T) {
	for _, s := range serverCache.backends {
		require(t, !s.running, "server shouldn't be running at start, but is", fmt.Sprintf("%+v", s))
		assert(t, !s.isRunning(), "isRunning method should reflect that server is not yet running")

		s.startIfNeeded()
		require(t, s.running, "server should be running after starting, but isn't", fmt.Sprintf("%+v", s))
		assert(t, s.isRunning(), "isRunning method should reflect that server is running")
		assert(t, s.pid != 0, "pid is set because server is running")
		cleanup(t, s)

		s = newServer(s.backend)
		s.addError(errors.New("blocker"))
		assert(t, s.hasError(), "error should be here")
		assert(t, !s.running, "server shouldn't start if it has errors")
		assert(t, !s.isRunning(), "server isn't running and shouldn't report that")
		s.startIfNeeded()
		assert(t, !s.running, "server shouldn't start if it has errors")
		assert(t, !s.isRunning(), "server isn't running and shouldn't report that")
		assert(t, s.pid == 0, "pid isnt set because server is running")

		cleanup(t, s)

		s = newServer(s.backend)
		s.running = true
		assert(t, s.isRunning(), "test faked running attribute and the method should reflect that")
		assert(t, !s.hasError(), "no errrors")
		s.startIfNeeded()
		assert(t, s.pid == 0, "pid isnt set because server is running")
		assert(t, !s.hasError(), "no errrors")

		s.setup()
	}
}

func TestErrorConditionsWhenStarting(t *testing.T) {
	for _, s := range serverCache.backends {
		wd := s.workingDirectory
		_, err := os.Stat(wd)
		assert(t, !os.IsNotExist(err), "working directory should be created by constructor:", wd)

		s.workingDirectory = s.workingDirectory + "-DOES-NOT-EXSIT"
		assert(t, s.workingDirectory != wd, "post modification wd shouldn't be correct")

		_, err = os.Stat(s.workingDirectory)
		assert(t, os.IsNotExist(err), "invalid working directory shouldn't exist:", s.workingDirectory)

		s.start()
		assert(t, s.hasError(), "server should have error after starting with a broken working directory", s.getError())
		assert(t, !s.isRunning(), "server shouldn't think it's running when it fails to start")

		s.workingDirectory = wd // set things right so that cleanup happens

		cleanup(t, s)
	}
}

func TestStartStop(t *testing.T) {
	for _, s := range serverCache.backends {
		s.start()
		assert(t, s.isRunning(), "server should run after its started")

		s.stop()
		assert(t, s.hasTerminated(), "server should report as terminated after it has")
		assert(t, !s.isRunning(), "server shouldn't be running  after its started")

		cleanup(t, s)
	}
}

func TestServersStopIsSafeToRunAfterCleanup(t *testing.T) {
	for _, s := range serverCache.backends {
		s.start()
		assert(t, s.isRunning(), "server should run after its started")

		s.stop()
		assert(t, s.hasTerminated(), "server knows it's terminated (0)")
		assert(t, !s.isRunning(), "server shouldn't be running  after its started (0)")
		s.stop()
		assert(t, s.hasTerminated(), "server knows it's terminated (1)")
		assert(t, !s.isRunning(), "server shouldn't be running  after its started (1)")

		cleanup(t, s)
	}
}
