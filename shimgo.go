package shimgo

import (
	"fmt"
)

func Cleanup()                                           { serverCache.cleanup() }
func Reset()                                             { serverCache.reset() }
func SupportsRst() bool                                  { _, ok := serverCache.getServer(RST); return ok }
func SupportsAsciiDoc() bool                             { _, ok := serverCache.getServer(ASCIIDOC); return ok }
func ConvertFromRst(content []byte) ([]byte, error)      { return convertHelper(RST, content) }
func ConvertFromAsciiDoc(content []byte) ([]byte, error) { return convertHelper(ASCIIDOC, content) }

func convertHelper(f Format, content []byte) ([]byte, error) {
	server, ok := serverCache.getServer(f)
	if !ok {
		return nil, fmt.Errorf("No suitable backend was found for %s", f)
	}

	return server.doConversion(f, content)
}
