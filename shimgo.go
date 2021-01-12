package shimgo

import (
	"fmt"
)

func Cleanup()                                           { serverCache.cleanup() }
func Reset()                                             { serverCache.reset() }
func SupportsRst() bool                                  { return serverCache.hasSupport(RST) }
func SupportsAsciiDoc() bool                             { return serverCache.hasSupport(ASCIIDOC) }
func SupportsAsciidoctor() bool                          { return serverCache.hasSupport(ASCIIDOCTOR) }
func ConvertFromRst(content []byte) ([]byte, error)      { return convertHelper(RST, content) }
func ConvertFromAsciiDoc(content []byte) ([]byte, error) { return convertHelper(ASCIIDOC, content) }
func ConvertFromAsciidoctor(content []byte) ([]byte, error) {
	return convertHelper(ASCIIDOCTOR, content)
}

func convertHelper(f Format, content []byte) ([]byte, error) {
	server, err := serverCache.getServer(f)
	if err != nil {
		return nil, fmt.Errorf("no suitable backend for '%s' was found: %s", f, err.Error())
	}

	return server.doConversion(f, content)
}
