package cli

import (
	"fmt"
	"time"
)

type Spinner struct {
	chars  []rune
	stopCh chan struct{}
	msg    string
}

func NewSpinner(msg string) *Spinner {
	return &Spinner{
		chars:  []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'},
		stopCh: make(chan struct{}),
		msg:    msg,
	}
}

func (s *Spinner) Start() {
	go func() {
		i := 0
		for {
			select {
			case <-s.stopCh:
				return
			default:
				fmt.Printf("\r%s %c", s.msg, s.chars[i%len(s.chars)])
				time.Sleep(100 * time.Millisecond)
				i++
			}
		}
	}()
}

func (s *Spinner) Stop(finalMsg string) {
	close(s.stopCh)
	// Clear the line before printing final message
	fmt.Printf("\r%-60s\n", "")
	fmt.Printf("%s [OK]\n", finalMsg)
}
