package parser

//
// AST
//

type Expr interface{}

// Types

type Char struct {
	Value byte
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

type Lambda struct {
	// name is auto-generated when we encounter the lambda
	Name string

	Params []string
	Exprs  []Expr

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
