package main

import (
	"io"
	"testing"
	"time"
)

func TestExitWhenClosedWaitsForTheParentToClose(t *testing.T) {
	r, w := io.Pipe()
	exited := make(chan struct{})
	go exitWhenClosed(r, func() { close(exited) })

	// Data from the parent is not a reason to stop.
	if _, err := w.Write([]byte("keepalive")); err != nil {
		t.Fatal(err)
	}
	select {
	case <-exited:
		t.Fatal("exited while the parent still holds the pipe open")
	case <-time.After(50 * time.Millisecond):
	}

	w.Close() // the parent exits: the OS closes its end
	select {
	case <-exited:
	case <-time.After(2 * time.Second):
		t.Fatal("did not exit after the pipe closed")
	}
}
