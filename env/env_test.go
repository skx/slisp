package env

import (
	"testing"
)

// TestBasic runs trivial/basic test of functionality
func TestBasic(t *testing.T) {

	// create env
	e := New(nil)

	// should be empty
	if e.CountLocals() != 0 {
		t.Fatalf("empty env had local variables, which is wrong")
	}

	// but define a new value and that's oka
	e.Define("foo", 3)
	e.DefineCapture("cat", 1)
	if e.CountLocals() != 1 {
		t.Fatalf("incorrect local-count")
	}

	// create a child
	c := New(e)
	c.Define("meow", 3)
	c.DefineCapture("dog", 391)

	// We now have two slots used.
	out := c.Names()
	if len(out) != 2 {
		t.Fatalf("wrong count of names")
	}

	// test we can lookup from both parent and child
	_, ok1 := c.Lookup("meow")
	if !ok1 {
		t.Fatalf("failed to lookup")
	}
	_, ok2 := c.Lookup("foo")
	if !ok2 {
		t.Fatalf("failed to lookup")
	}
	_, ok3 := c.Lookup("foo.not.found")
	if ok3 {
		t.Fatalf("lookup resulted success on a missing variable")
	}

	// test we can lookup from both parent and child
	_, ok4 := c.LookupCapture("dog")
	if !ok4 {
		t.Fatalf("failed to lookup")
	}
	_, ok5 := c.LookupCapture("cat")
	if !ok5 {
		t.Fatalf("failed to lookup")
	}
	_, ok6 := c.LookupCapture("foo.not.found")
	if ok6 {
		t.Fatalf("lookup resulted success on a missing variable")
	}

}
