package util

import (
	"bufio"
	"os"
	"testing"
)

func TestInterrupt(t *testing.T) {
	stderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Error(err)
	}
	os.Stderr = w
	defer func() {
		os.Stderr = stderr
		_ = w.Close()
	}()

	scanner := bufio.NewScanner(r)
	i := NewInterrupt()
	defer i.Stop()

	for j := 0; j < 10; j++ {
		i.signal <- os.Interrupt
		scanner.Scan()
		if scanner.Text() != "Current execution cannot be interrupted!" {
			t.Errorf("unexpected behavior of interrupt blocker")
		}
	}
}
