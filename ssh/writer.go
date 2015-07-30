package ssh

import (
	"fmt"
	"io"

	"golang.org/x/crypto/ssh/terminal"
)

// ResponseWriter writes data and status codes to the client
type ResponseWriter struct {
	Colors *terminal.EscapeCodes
	writer io.WriteCloser
}

func (r *ResponseWriter) colorCode(color []byte, code StatusCode, format string, args ...interface{}) {

	// Set color
	r.writer.Write(color)

	// Write error name and code
	if t, ok := statusCodes[code]; ok {
		r.writer.Write([]byte(fmt.Sprintf(" %s (%d)", t, int(code))))
	} else {
		r.writer.Write([]byte(fmt.Sprintf(" Unknown (%d)", int(code))))
	}

	// Write the error message if there is one
	if len(format) > 0 {
		r.writer.Write([]byte(": "))
		fmt.Fprintf(r.writer, format, args...)
	}

	// Reset terminal colors
	r.writer.Write(r.Colors.Reset)
	r.writer.Write([]byte("\r\n"))
}

// Fail writes the error status code to the writer
func (r *ResponseWriter) Fail(code StatusCode, format string, args ...interface{}) {
	r.colorCode(r.Colors.Red, code, format, args...)
}

// Success writes the status code to the writer
func (r *ResponseWriter) Success(code StatusCode, format string, args ...interface{}) {
	r.colorCode(r.Colors.Green, code, format, args...)
}

// Write is a pass through function into the underlying writer
func (r *ResponseWriter) Write(data []byte) (int, error) {
	return r.writer.Write(data)
}
