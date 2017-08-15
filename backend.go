package shimgo

import (
	"errors"
	"os/exec"
	"path/filepath"
)

type backend string

const (
	pythonServer backend = "http://localhost:1414/"
)

func (b backend) writeFiles(workingDirectory string) error {
	switch b {
	case pythonServer:
		return writePythonFiles(workingDirectory)
	default:
		return errors.New("unsupported backend")
	}
}

func (b backend) getCommand(workingDirectory string) *exec.Cmd {
	switch b {
	case pythonServer:
		return exec.Command(getPython2(), filepath.Join(workingDirectory, pythonService))
	default:
		return nil
	}
}
