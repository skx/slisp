package parser

//
// AST
//

type Expr any

// Types

type Char struct {
	Value byte
}

type Float struct {
	Value float64
}

type Int struct {
	Value int64
}

type String struct {
	Value string
}

type Symbol struct {
	Name string
}

type Nil struct {
}

// specials

// TopLevel is an interface which must be satisfied by an expression which
// is valid to appear at the top-level of user programs.
//
// In the past we only allowed function definitions at the top-level, but
// now we allow global variable/constant definitions too.  This interface
// must be implemented by something which wants to be valid at the top-level.
type TopLevel interface {
	Type() string
}

type Binding struct {
	Name string
	Expr Expr
}

type Call struct {
	Fn   Expr
	Args []Expr
}

type CondCase struct {
	Case  Expr
	Exprs []Expr
}

type Cond struct {
	Cases []CondCase
}

// Defun holds a function definition.
//
// As this implements the TopLevel interface it may appear at the top-level of a slisp file.
type Defun struct {
	// Name of the function being defined.
	Name string

	// The names of the parameter variables.
	Params []string

	// Is this function variadic?
	// If so the last argument will get a list.
	Variadic bool

	// Exprs contains the expressions in the body of the function.
	Exprs []Expr
}

// Type is the implementation of the TopLevel interface
func (d Defun) Type() string { return "defun" }

type Do struct {
	Exprs []Expr
}

type If struct {
	Cond Expr
	Then Expr
	Else Expr
}

// Global is used to declare a global variable, or constant.
//
// As this implements the TopLevel interface it may appear at the top-level of a slisp file.
type Global struct {
	// Const is true for read-only variables
	Const bool

	// Init records if the variable has been given its initial value
	Init bool

	// Name is the name of the global variable.
	Name string

	// Value will evaluate to the value.
	Value Expr
}

// Type is the implementation of the TopLevel interface
func (d Global) Type() string { return "global" }

// Lambda represents a lambda, which is basically identical to a Defun.
// The only difference is a list of captured variables, so we'll embed
// the Defun and treat it as one most of the time.
type Lambda struct {
	Defun

	// Captured variables - we don't do free-variable analysis,
	// and just capture all the variables we could.
	Captures []string
}

type Let struct {
	Bindings []Binding
	Body     []Expr
}

type Set struct {
	Name string
	Expr Expr
}

type While struct {
	Cond  Expr
	Exprs []Expr
}
