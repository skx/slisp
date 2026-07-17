// Package env stores our scopes of valid environmental symbols, and
// their offsets against the stack-pointer, or register used for
// lambdas.
package env

import "fmt"

// Env holds our state
type Env struct {
	// parent, if any
	parent *Env

	// slots are for local variables, relative to RBP.
	slots map[string]int

	// order lets us keep order, since calling range over a map
	// will not return the same order as insertion time.
	order []string

	// captures hold offsets against R15 for captured
	// variables inside closures.
	captures map[string]int

	// shared by all child environments in a function, used to denote the
	// number of stack locals for arguments.
	maxOffset *int

	// shared by all child environments in a function; used to
	// generate unique names for compiler-generated temporary slots.
	tempCounter *int
}

// New creates a new environment, with an optional parent.
func New(parent *Env) *Env {
	var maxOffset *int
	var tempCounter *int

	if parent != nil {
		maxOffset = parent.maxOffset
		tempCounter = parent.tempCounter
	} else {
		maxOffset = new(int)
		tempCounter = new(int)
	}

	return &Env{
		parent:      parent,
		slots:       map[string]int{},
		captures:    map[string]int{},
		maxOffset:   maxOffset,
		tempCounter: tempCounter,
	}
}

// NewTemp reserves a stack slot for a compiler-generated temporary value
// relative to RBP, exactly like Define.
func (e *Env) NewTemp() int {
	name := fmt.Sprintf(" tmp%d", *e.tempCounter)
	*e.tempCounter++
	return e.Define(name)
}

// Define defines a local variable, and returns the offset relative to the RBP register.
func (e *Env) Define(name string) int {

	offset := (e.countLocals() + 2) * 8
	e.slots[name] = offset
	e.order = append(e.order, name)

	if offset > *e.maxOffset {
		*e.maxOffset = offset
	}
	return offset
}

// MaxOffset returns the maximum offset for a stack-local argument.
func (e *Env) MaxOffset() int {
	return *e.maxOffset
}

// DefineCapture defines a captured variable, in this case the offset returned will be used
// relative to the R15 closure-base register.
func (e *Env) DefineCapture(name string) int {
	offset := (len(e.captures) + 1) * 8
	e.captures[name] = offset
	return offset
}

// countLocals returns the number of local variables defined in this,
// and any parent scopes.  It is necessary to calculate the offset to
// use for stack-local addressing.
func (e *Env) countLocals() int {
	used := len(e.slots)
	if e.parent != nil {
		used += e.parent.countLocals()
	}
	return used
}

// Names returns all the names of variables known at this level,
// and all parent levels.
//
// We use this as a hack for lambda-closures, instead of performing
// real free-variable analysis.
func (e *Env) Names() []string {
	var out []string
	if e.parent != nil {
		out = append(out, e.parent.Names()...)
	}
	out = append(out, e.order...)
	return out
}

// LookupCapture performs the same lookup function for lambdas,
// as part of our closure implementation.
func (e *Env) LookupCapture(name string) (int, bool) {

	if v, ok := e.captures[name]; ok {
		return v, true
	}

	if e.parent != nil {
		return e.parent.LookupCapture(name)
	}
	return 0, false
}

// Lookup returns the slot-index of the given variable-name.
//
// If not found in the current scope the parent(s) will be searched, recursively.
func (e *Env) Lookup(name string) (int, bool) {

	if v, ok := e.slots[name]; ok {
		return v, true
	}

	if e.parent != nil {
		return e.parent.Lookup(name)
	}

	return 0, false
}
