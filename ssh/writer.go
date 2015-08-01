package ssh

import (
	"fmt"
	"io"
)

type ColorCodes struct {
	// Foreground colors
	Black, Red, Green, Yellow, Blue, Magenta, Cyan, White []byte

	// Light Foreground colors
	LightGrey, LightRed, LightGreen, LightYellow, LightBlue, LightMagenta, LightCyan, LightWhite []byte

	// Reset all attributes
	Reset []byte
}

const keyEscape = 27

// ANSI colors
var DefaultColorCodes = ColorCodes{
	Black:   []byte{keyEscape, '[', '3', '0', 'm'},
	Red:     []byte{keyEscape, '[', '3', '1', 'm'},
	Green:   []byte{keyEscape, '[', '3', '2', 'm'},
	Yellow:  []byte{keyEscape, '[', '3', '3', 'm'},
	Blue:    []byte{keyEscape, '[', '3', '4', 'm'},
	Magenta: []byte{keyEscape, '[', '3', '5', 'm'},
	Cyan:    []byte{keyEscape, '[', '3', '6', 'm'},
	White:   []byte{keyEscape, '[', '3', '7', 'm'},

	LightGrey:    []byte{keyEscape, '[', '9', '0', 'm'},
	LightRed:     []byte{keyEscape, '[', '9', '1', 'm'},
	LightGreen:   []byte{keyEscape, '[', '9', '2', 'm'},
	LightYellow:  []byte{keyEscape, '[', '9', '3', 'm'},
	LightBlue:    []byte{keyEscape, '[', '9', '4', 'm'},
	LightMagenta: []byte{keyEscape, '[', '9', '5', 'm'},
	LightCyan:    []byte{keyEscape, '[', '9', '6', 'm'},
	LightWhite:   []byte{keyEscape, '[', '9', '7', 'm'},
	Reset:        []byte{keyEscape, '[', '0', 'm'},
}

// ResponseWriter writes data and status codes to the client
type ResponseWriter struct {
	Colors ColorCodes
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
	r.colorCode(r.Colors.LightRed, code, format, args...)
}

// Success writes the status code to the writer
func (r *ResponseWriter) Success(code StatusCode, format string, args ...interface{}) {
	r.colorCode(r.Colors.LightGreen, code, format, args...)
}

// Write is a pass through function into the underlying writer
func (r *ResponseWriter) Write(data []byte) (int, error) {
	return r.writer.Write(data)
}
