package shimgo

import (
	"errors"
	"fmt"
)

var servers []*shimServer

func init() {
	servers = append(servers, newServer(PYTHON))
}

func Cleanup() {
	for _, server := range servers {
		server.stop()
	}
}

func Reset() {
	for _, server := range servers {
		wasRunning := server.isRunning()

		server.stop()
		server.reset()

		if wasRunning {
			server.start()
		}
	}
}

func ConvertFromRst(content []byte) ([]byte, error) {
	for _, server := range servers {
		if err := server.startIfNeeded(); err != nil {
			return nil, err
		}

		if server.supportsConversion(RST) {
			return server.doConversion(RST, content)
		}
	}
	return nil, errors.New(fmt.Sprintf("No suitable backend was found for %s", RST))
}

func ConvertFromAsciiDoc(content []byte) ([]byte, error) {
	for _, server := range servers {
		if err := server.startIfNeeded(); err != nil {
			return nil, err
		}

		if server.supportsConversion(ASCIIDOC) {
			return server.doConversion(ASCIIDOC, content)
		}
	}
	return nil, errors.New(fmt.Sprintf("No suitable backend was found for %s", ASCIIDOC))
}

func SupportsRst() bool {
	for _, server := range servers {
		if err := server.startIfNeeded(); err != nil {
			return false
		}

		return server.supportsConversion(RST)
	}
	return false
}

func SupportsAsciiDoc() bool {
	for _, server := range servers {
		if err := server.startIfNeeded(); err != nil {
			return false
		}

		return server.supportsConversion(ASCIIDOC)
	}
	return false
}
