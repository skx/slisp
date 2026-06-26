package env

import (
	"testing"
)

// TestBasic runs trivial/basic test of functionality
func TestBasic(t *testing.T) {

	// create env
	e := New(nil)

	// will be empty - define things
	e.Define("foo")
	e.DefineCapture("cat")

	// create a child
	c := New(e)
	c.Define("meow")
	c.DefineCapture("dog")

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
