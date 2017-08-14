package shimgo

func Cleanup() {
	server.stop()
}

func Reset() {
	wasRunning := server.isRunning()

	server.stop()
	server = newServer()

	if wasRunning {
		server.start()
	}
}

func ConvertFromRst(content []byte) ([]byte, error) {
	if err := server.startIfNeeded(); err != nil {
		return nil, err
	}

	return server.doConversion(RST, content)
}

func ConvertFromAsciiDoc(content []byte) ([]byte, error) {
	if err := server.startIfNeeded(); err != nil {
		return nil, err
	}

	return server.doConversion(ASCIIDOC, content)
}

func SupportsAsciiDoc() bool {
	if err := server.startIfNeeded(); err != nil {
		return false
	}

	return server.supportsConversion(ASCIIDOC)
}

func SupportsRst() bool {
	if err := server.startIfNeeded(); err != nil {
		return false
	}

	return server.supportsConversion(RST)
}
