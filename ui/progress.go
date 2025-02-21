package ui

import (
	"fmt"
	"sync"
	"time"
)

var spinnerChars = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type Spinner struct {
	stop    chan struct{}
	stopped bool
	sync.Mutex
}

func NewSpinner() *Spinner {
	return &Spinner{
		stop: make(chan struct{}),
	}
}

func (s *Spinner) Start(message string) {
	s.Lock()
	if s.stopped {
		s.Unlock()
		return
	}
	s.Unlock()

	go func() {
		for i := 0; ; i = (i + 1) % len(spinnerChars) {
			s.Lock()
			if s.stopped {
				s.Unlock()
				return
			}
			s.Unlock()

			// Clear line and print spinner
			fmt.Printf("\r\033[K%s %s", spinnerChars[i], message)

			select {
			case <-s.stop:
				return
			case <-time.After(100 * time.Millisecond):
				continue
			}
		}
	}()
}

func (s *Spinner) Stop() {
	s.Lock()
	defer s.Unlock()
	if s.stopped {
		return
	}
	s.stopped = true
	close(s.stop)
	// Clear the line
	fmt.Print("\r\033[K")
}
