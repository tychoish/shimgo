package shimgo

import (
	"errors"
	"os/exec"
	"path/filepath"
)

type backend int

const (
	pythonServer backend = iota
	rubyServer
)

func (b backend) writeFiles(workingDirectory string) error {
	switch b {
	case pythonServer:
		return writeFiles([]string{pythonService, asciidoc, asciidocapi}, workingDirectory)
	case rubyServer:
		return writeFiles([]string{rubyService}, workingDirectory)
	default:
		return errors.New("unsupported backend")
	}
}

func (b backend) getCommand(workingDirectory, port string) *exec.Cmd {
	switch b {
	case pythonServer:
		return exec.Command(getPython2(), filepath.Join(workingDirectory, pythonService), port)
	case rubyServer:
		return exec.Command(getRuby(), filepath.Join(workingDirectory, rubyService), port)
	default:
		return nil
	}
}
