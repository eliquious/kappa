package ssh

import (
	"fmt"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

type ResponseWriter struct {
	Colors  *terminal.EscapeCodes
	channel ssh.Channel
}

func (r *ResponseWriter) colorCode(color []byte, code StatusCode, format string, args ...interface{}) {
	r.channel.Write(color)

	if t, ok := statusCodes[code]; ok {
		r.channel.Write([]byte(fmt.Sprintf(" %s (%d)", t, int(code))))
	} else {
		r.channel.Write([]byte(fmt.Sprintf(" Unknown (%d)", int(code))))
	}

	if len(format) > 0 {
		r.channel.Write([]byte(" : "))
		fmt.Fprintf(r.channel, format, args...)
	}
	r.channel.Write(r.Colors.Reset)
	r.channel.Write([]byte("\r\n"))
}

func (r *ResponseWriter) Fail(code StatusCode, format string, args ...interface{}) {
	r.colorCode(r.Colors.Red, code, format, args...)
}

func (r *ResponseWriter) Success(code StatusCode, format string, args ...interface{}) {
	r.colorCode(r.Colors.Green, code, format, args...)
}

func (r *ResponseWriter) Write(data []byte) (int, error) {
	return r.channel.Write(data)
}
