package parser

//
// AST
//

type Expr interface{}

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

type Defun struct {
	Name   string
	Params []string
	Exprs  []Expr
}

type Do struct {
	Exprs []Expr
}

type If struct {
	Cond Expr
	Then Expr
	Else Expr
}

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

