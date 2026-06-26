// Package env stores our scopes of valid environmental symbols, and
// their offsets against the stack-pointer, or register used for
// lambdas.
package env

// Env holds our state
type Env struct {
	// parent, if any
	parent *Env

	// slots are for local variables, relative to RBP.
	slots map[string]int

	// captures hold offsets against R15 for captured
	// variables inside closures.
	captures map[string]int
}

// New creates a new environment, with an optional parent.
func New(parent *Env) *Env {
	return &Env{
		parent:   parent,
		slots:    map[string]int{},
		captures: map[string]int{},
	}
}

// Define defines a local variable, in this case
// the offset is relative to the RBP register.
func (e *Env) Define(name string, offset int) {
	e.slots[name] = offset
}

// DefineCapture defines a captured variable, in this case
// the offset is relative to the R15 closure-base register.
func (e *Env) DefineCapture(name string, offset int) {
	e.captures[name] = offset
}

func (e *Env) CountLocals() int {
	ours := len(e.slots)
	if e.parent != nil {
		ours += e.parent.CountLocals()
	}
	return ours
}

// Names returns all the names of variables known at this level,
// and all parent levels.
//
// We use this as a hack for lambda-closures, instead of performing
// real free-variable analysis.
func (e *Env) Names() []string {
	var res []string

	for k := range e.slots {
		res = append(res, k)
	}
	if e.parent != nil {
		parents := e.parent.Names()
		res = append(res, parents...)
	}
	return res
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
