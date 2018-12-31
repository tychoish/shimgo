package shimgo

import (
	"errors"
	"os/exec"
	"path/filepath"
)

type backend string

const (
	pythonServer backend = "http://localhost:1414"
	rubyServer           = "http://localhost:1515"
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

func (b backend) getCommand(workingDirectory string) *exec.Cmd {
	switch b {
	case pythonServer:
		return exec.Command(getPython2(), filepath.Join(workingDirectory, pythonService))
	case rubyServer:
		return exec.Command(getRuby(), filepath.Join(workingDirectory, rubyService))
	default:
		return nil
	}
}
