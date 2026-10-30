package search

import (
	"errors"
	"net"
	"testing"
)

func TestIsEisAccessError(t *testing.T) {
	if !isEisAccessError(&net.OpError{Op: "dial", Err: errors.New("connection refused")}) {
		t.Fatal("expected access error")
	}
	if !isEisAccessError(errors.New("EIS 44 SOAP URL not configured")) {
		t.Fatal("expected not configured")
	}
	if isEisAccessError(errors.New("validation failed")) {
		t.Fatal("unexpected access error")
	}
}
