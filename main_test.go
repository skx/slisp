package main

import (
	"strings"
	"testing"
)

func TestOK(t *testing.T) {

	out, err := compile("(defun main() (print 22))")
	if err != nil {
		t.Fatalf("unexpected error compiling: %s", err)
	}
	if !strings.Contains(out, "call fn_main") {
		t.Fatalf("compiled version looks suspect")
	}
}
