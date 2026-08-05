package parser

//
// AST
//

// Expr is the catchall expression type.
type Expr any

// Basic/Literal Types

// Char holds a character literal.
type Char struct {
	Value byte
}

// Float holds a floating point literal.
type Float struct {
	Value float64
}

// Int holds an integer literal.
type Int struct {
	Value int64
}

// String holds a string literal.
type String struct {
	Value string
}

// Symbol holds a symbol name.
type Symbol struct {
	Name string
}

// Nil holds a nil-literal.
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

// Alias allows rewriting functions, and is designed to allow packages to override
// the default behaviour provided by our standard-library.
type Alias struct {
	Old string
	New string
}

// Type is the implementation of the TopLevel interface.
func (d Alias) Type() string { return "alias" }

// Binding holds the name and value of variables within a new scope, started with Let.
type Binding struct {
	Name string
	Expr Expr
}

// Call is used to represent function-calls.
type Call struct {
	Fn   Expr
	Args []Expr
}

// Defmacro holds a macro definition.
type Defmacro struct {
	// Name of the macro being defined.
	Name string

	// The names of the parameter variables.
	Params []string

	// Is this macro variadic?
	// If so the last argument will be bound to a List of the
	// remaining, literal, argument-expressions.
	Variadic bool

	// Exprs contains the expressions in the body of the macro.
	Exprs []Expr
}

// Type is the implementation of the TopLevel interface.
func (d Defmacro) Type() string { return "defmacro" }

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

// Type is the implementation of the TopLevel interface.
func (d Defun) Type() string { return "defun" }

// Do allows running multiple expressions in a context where only a single one is permitted.
type Do struct {

	// Exprs contains the list of expressions to execute.
	Exprs []Expr
}

// If is our conditional expression.
type If struct {
	// Cond holds the test to make.
	Cond Expr

	// Then is the expression executed if the test passes.
	Then Expr

	// Else (optional) holds the expression executed if the test fails.
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

// Type is the implementation of the TopLevel interface.
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

// Let introduces a new scope, with the given binings, then executes the named Body.
type Let struct {
	Bindings []Binding
	Body     []Expr
}

// List represents a literal sequence of expressions.
//
// This is distinct from a real list, and used internally for parsing.
type List struct {
	Elems []Expr
}

// Quote represents a quoted expression - 'expr - which prevents
// evaluation of expr.
type Quote struct {
	Expr Expr
}

// Quasiquote represents a quasiquoted expression - `expr - which
// behaves like Quote except that any Unquote/UnquoteSplicing nested
// within it are evaluated (or substituted, within a macro body) and
// spliced into the result.
type Quasiquote struct {
	Expr Expr
}

// Require allows loading code from another file, be it an embedded
// package, or a user-provided one.
type Require struct {
	Feature string
}

// Type is the implementation of the TopLevel interface.
func (r Require) Type() string { return "require" }

// Set allows storing the result of an expression in a variable with the given name.
type Set struct {
	Name string
	Expr Expr
}

// Unquote represents ",expr" - only valid when nested within a Quasiquote.
type Unquote struct {
	Expr Expr
}

// UnquoteSplicing represents ",@expr" - only meaningful as a list
// element nested within a Quasiquote.  The value of Expr is expected to
// be a list, whose elements are spliced into the surrounding list.
type UnquoteSplicing struct {
	Expr Expr
}

// While holds a loop construct.
type While struct {
	Cond  Expr
	Exprs []Expr
}
