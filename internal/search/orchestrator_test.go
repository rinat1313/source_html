package search

import (
	"errors"
	"net"
	"testing"
)

func TestIsTransportError(t *testing.T) {
	if !isTransportError(&net.OpError{Op: "dial", Err: errors.New("connection refused")}) {
		t.Fatal("expected transport")
	}
	if isTransportError(errors.New("validation failed")) {
		t.Fatal("unexpected transport")
	}
}
