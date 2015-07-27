package ssh

import (
	"crypto/rand"
	"io"
	"math/big"

	"golang.org/x/crypto/ssh/terminal"
)

type Message interface {
	Write(w io.Writer, colors terminal.EscapeCodes)
}

type SimpleMessage struct {
	Text string
}

func (s SimpleMessage) Write(w io.Writer, colors terminal.EscapeCodes) {
	w.Write([]byte(" " + s.Text + "\r\n"))
}

// Quote is a message with an author
type Quote struct {
	Color  []byte
	Author string
	Text   string
}

func (q Quote) Write(w io.Writer, colors terminal.EscapeCodes) {
	w.Write(q.Color)
	w.Write([]byte(" " + q.Author))
	w.Write(colors.Reset)
	w.Write([]byte(": " + q.Text))
	w.Write([]byte("\r\n"))
}

type Conversation struct {
	Quotes []Quote
}

func (c Conversation) Write(w io.Writer, colors terminal.EscapeCodes) {
	for _, q := range c.Quotes {
		q.Write(w, colors)
	}
	w.Write([]byte("\r\n"))
}

func LoginMessage(w io.Writer, colors terminal.EscapeCodes) {
	var messages = []Message{
		SimpleMessage{"Welcome to Kappa DB, Yo!"},
		Quote{colors.Magenta, "Jessy Pinkman", "Yeah, Bitch! Magnets!"},
		Quote{colors.Magenta, "Jessy Pinkman", "Yeah, Science!"},
		Quote{colors.Blue, "Saul Goodman", "Better call Saul."},
		Conversation{
			Quotes: []Quote{
				Quote{colors.Green, "Walter White", "One particular element comes to mind..."},
				Quote{colors.Magenta, "Jessy Pinkman", "Ohhhhh... wire.."},
			},
		},
	}

	// messages := []string{
	// 	"One particular element comes to mind... Ahhhh.. wire..",
	// }

	index, err := rand.Int(rand.Reader, big.NewInt(int64(len(messages))))
	if err != nil {
		index = big.NewInt(0)
	}
	messages[index.Int64()].Write(w, colors)
}
