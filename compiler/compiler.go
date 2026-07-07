// Package compiler is our main workhorse, which creates an assembly
// language version of the given input program and outputs it to STDOUT.
package compiler

import (
	"bytes"
	"crypto/sha1"
	_ "embed"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/skx/slisp/env"
	"github.com/skx/slisp/parser"
)

//go:embed template.tmpl
var tmplTxt string

// FunctionArgs records the functions that defuns accept.
//
// We need this because we need to discover how many arguments
// each function expects - so we can abort if a function is
// called with the wrong number of arguments - and also to know
// if variable arguments are in-use.
type FunctionArgs struct {
	// How many arguments does this function expect?
	Arguments int

	// Is this a variadic function?
	Variadic bool
}

// Compiler holds our state
type Compiler struct {

	// source stores the program we're parsing.
	source string

	// text stores the text we emit as we compile various things.
	text strings.Builder

	// labelID is used to give unique labels to if/lambda/etc.
	labelID int

	// loaded contains packages we've already loaded
	loaded map[string]bool

	// strings holds the strings we've encountered, indexed
	// by their SHA1 sum as ID.  This is how we intern.
	strings map[string]string

	// floats holds the literal floating point numbers  we've encountered,
	// indexed by their SHA1 sum as ID.  This is how we intern.
	floats map[string]float64

	// lambdas holds the lambdas we've encountered.
	lambdas []*parser.Lambda

	// functions stores details about our defined functions, specifically
	// whether each one is variadic.
	functions map[string]*FunctionArgs

	// globals stores details of top-level global variables
	globals map[string]parser.Global

	// inPackage stores the name of the package we're currently inside, if any
	inPackage string
}

// New is our constructor
func New(src string) *Compiler {

	// return a new object, with the source and
	// all internal maps created.
	return &Compiler{
		source:    src,
		floats:    map[string]float64{},
		functions: map[string]*FunctionArgs{},
		globals:   map[string]parser.Global{},
		loaded:    map[string]bool{},
		strings:   map[string]string{},
	}
}

// findPackage tries to find the location from which to
// load .lisp files via "(require foo)"
func (c *Compiler) findPackage(file string) (string, error) {

	// Present in the CWD?
	if _, err := os.Stat(file); err == nil {
		return file, nil
	}

	// Otherwise search the path
	if path := os.Getenv("LISP_PATH"); path != "" {

		for dir := range strings.SplitSeq(path, ":") {

			if dir == "" {
				dir = "."
			}
			candidate := filepath.Join(dir, file)

			if _, err := os.Stat(candidate); err == nil {
				return candidate, nil
			}
		}
	}

	return "", fmt.Errorf("unable to locate package %q", file)
}

func (c *Compiler) expandRequires(defs []parser.TopLevel) ([]parser.TopLevel, error) {

	var out []parser.TopLevel

	for _, expr := range defs {

		r, ok := expr.(parser.Require)
		if !ok {
			out = append(out, expr)
			continue
		}

		name := r.Feature

		// Already loaded?
		if c.loaded[name] {
			continue
		}
		c.loaded[name] = true

		// If there is no suffix then add ".lisp"
		file := name
		if filepath.Ext(file) == "" {
			file += ".lisp"
		}

		path, err := c.findPackage(file)
		if err != nil {
			return nil, err
		}

		src, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		p := parser.New(string(src))

		pkg, err := p.Parse()
		if err != nil {
			return nil, err
		}

		// Expand any nested requires.
		pkg, err = c.expandRequires(pkg)
		if err != nil {
			return nil, err
		}

		// Ignore package declarations.
		for _, x := range pkg {
			if _, ok := x.(parser.Require); ok {
				continue
			}
			out = append(out, x)
		}
	}

	return out, nil
}

// Compile creates and returns the assembly language source for the given
// list of functions.
func (c *Compiler) Compile() (string, error) {

	// Create a parser
	p := parser.New(c.source)

	// Parse the program into functions
	defs, err := p.Parse()
	if err != nil {
		return "", fmt.Errorf("error parsing program %s", err)
	}

	defs, err = c.expandRequires(defs)
	if err != nil {
		return "", err
	}
	main := false

	//
	// Process each known function, and record the number
	// of arguments it requests, and whether the last argument
	// should be treated as variadic.
	//
	for _, fun := range defs {

		// Record global variables.
		g, ok1 := fun.(parser.Global)
		if ok1 {
			c.globals[g.Name] = g
		}

		// Record data about defined functions
		d, ok2 := fun.(parser.Defun)
		if ok2 {
			c.functions[d.Name] = &FunctionArgs{
				Arguments: len(d.Params),
				Variadic:  d.Variadic,
			}
		}

		// But defuns might be in packages
		p, ok3 := fun.(parser.Package)
		if ok3 {

			c.inPackage = p.Name
			for _, fun := range p.Contents {

				// Record global variables.
				g, ok1 := fun.(parser.Global)
				if ok1 {
					name := p.Name + ":" + g.Name
					c.globals[name] = g
				}

				// Record data about defined functions
				d, ok2 := fun.(parser.Defun)
				if ok2 {
					name := p.Name + ":" + d.Name
					c.functions[name] = &FunctionArgs{
						Arguments: len(d.Params),
						Variadic:  d.Variadic,
					}
				}
			}
			c.inPackage = ""
		}

	}

	//
	// This whole function is messy, but in brief
	// we assemble stuff into an internal buffer "g.text"
	// and at various points we need to read the contents
	// of that assembly as a string and then reset the
	// buffer.
	//
	// This inline function does that.
	//
	getCompiled := func() string {
		txt := c.text.String()
		c.text.Reset()
		return txt
	}

	// Create a new environment for the global defun/defvar
	// statements - they can't really use it, but it is required.
	e := env.New(nil)

	// Generate the values of the global variables.
	for _, tl := range defs {

		// Is this a global variable?
		g, ok1 := tl.(parser.Global)
		if ok1 {
			err = c.emitExpr(g.Value, e)
			if err != nil {
				return "", err
			}

			// This variable has been set now.
			x := c.globals[g.Name]
			x.Init = true
			c.globals[g.Name] = x

			c.emitln(fmt.Sprintf("    mov [%s], rax ; %s", c.addThing("global", g.Name), g.Name))
		}

		// Packages can have variables too
		p, ok2 := tl.(parser.Package)
		if ok2 {
			c.inPackage = p.Name
			for _, tl = range p.Contents {
				g, ok1 := tl.(parser.Global)

				if ok1 {
					name := p.Name + ":" + g.Name
					err = c.emitExpr(g.Value, e)
					if err != nil {
						return "", err
					}

					// This variable has been set now.
					x := c.globals[name]
					x.Init = true
					c.globals[name] = x

					c.emitln(fmt.Sprintf("    mov [%s], rax ; package %s %s", c.addThing("global", name), p.Name, name))
				}
			}
			c.inPackage = ""
		}

	}

	//
	// Compiled code to setup the initial value of
	// each known defvar/defconst.
	//
	// To be inserted into our rendered template shortly.
	//
	initGlobals := getCompiled()

	//
	// Now generate the assembly for each known user-defined
	// function to our internal buffer.
	//
	for _, tl := range defs {

		// We only care about defuns
		d, ok2 := tl.(parser.Defun)
		if ok2 {
			if d.Name == "main" {
				main = true
			}
			err = c.emitCallable(d)
			if err != nil {
				return "", err
			}
			c.emitln("")
		}

		// But defuns might be in packages
		p, ok3 := tl.(parser.Package)
		if ok3 {
			c.inPackage = p.Name
			for _, tl := range p.Contents {

				d, ok := tl.(parser.Defun)
				if ok {

					// modify the name
					d.Name = p.Name + ":" + d.Name
					err = c.emitCallable(d)
					if err != nil {
						return "", err
					}
					c.emitln("")
				}
			}
			c.inPackage = ""
		}
	}

	if !main {
		return "", fmt.Errorf("there is no entry-point defined; we need a defun named 'main'")
	}

	//
	// Get the compiled functions
	//
	defuns := getCompiled()

	//
	// Compile each known lambda function.
	//
	for _, l := range c.lambdas {
		err = c.emitCallable(l)
		if err != nil {
			return "", err
		}

		c.emitln("")
	}

	//
	// Get their compiled bodies
	//
	lambdas := getCompiled()

	//
	// Build up a data-section for our string tables
	//
	c.emitln("section .data")
	for id, str := range c.strings {
		c.emitln("align 8")
		c.emitln(id + ":")

		// escape the "`" which are wrapped around the string.
		str = strings.ReplaceAll(str, "`", "\\`")

		c.emitln(fmt.Sprintf("     db `%s`, 0", str))
	}

	// Now as a simple string
	stringTable := getCompiled()

	//
	// Build up a data-section for our user-defined float
	// literals
	//
	c.emitln("section .data")
	for id, str := range c.floats {
		c.emitln("align 8")
		c.emitln(id + ":")
		c.emitln(fmt.Sprintf("     dq %f", str))
	}
	floatTable := getCompiled()

	//
	// We also need to define a variable to hold the pointer
	// for each global-variable value.
	//
	globals := []string{}
	for nm := range c.globals {
		globals = append(globals, c.addThing("global", nm))
	}

	//
	// Define a simple structure we can pass to the text/template
	// file we render for our output.
	//
	type Generated struct {
		// The defintions of defun's we've seen.
		Defuns string

		// Lambdas has all the lambda expressions we've seen.
		Lambdas string

		// InitGlobals is the thing that loads global variables
		InitGlobals string

		// Globals has global variables
		Globals []string

		// StringTable contains the strings we've seen.
		StringTable string

		// FloatTable contains the floating point literals we've seen.
		FloatTable string
	}

	//
	// Create an instance of that internal structure, which we
	// can then pass to the template processor to fill out into
	// the template appropriately.
	//
	x := &Generated{
		Defuns:      defuns,
		Globals:     globals,
		InitGlobals: initGlobals,
		Lambdas:     lambdas,
		StringTable: stringTable,
		FloatTable:  floatTable,
	}

	// Create a buffer to render the template to.
	buf := bytes.Buffer{}

	// Load the template, and parse it.
	t1 := template.New("t1")
	t1 = template.Must(t1.Parse(tmplTxt))

	// Render the template.
	err = t1.Execute(&buf, x)
	if err != nil {
		return "", err
	}

	// Now return the text of that rendered template.
	return buf.String(), nil
}

// addThing creates a unique label for our floats,
// and strings, based on the SHA1-hash.  Interning them.
func (c *Compiler) addThing(prefix string, f any) string {
	hasher := sha1.New()
	hasher.Write(fmt.Appendf(nil, "%v", f))
	sha := hex.EncodeToString(hasher.Sum(nil))
	id := fmt.Sprintf("%s_%s", prefix, sha)
	return id
}

// label generates a new label, with the given prefix.
func (c *Compiler) label(prefix string) string {
	s := fmt.Sprintf("%s_%d", prefix, c.labelID)
	c.labelID++
	return s
}

// emitln writes a line of assembly/source into our internal buffer.
func (c *Compiler) emitln(s string) {
	c.text.WriteString(s)
	c.text.WriteString("\n")
}

// asmName converts the given label into something nasm will accept.
//
// It doesn't like special characters inside label names, and compiling
// a function with a name like "not" or "abs" will cause errors when
// they're called.  ("call abs" will result in a syntax error from nasm.)
func (c *Compiler) asmName(name string) string {

	switch name {

	// comparisons
	case "=":
		return "equals"
	case "!":
		return "fn_not"
	case "<=":
		return "lt_equals"
	case "<":
		return "lt"
	case ">":
		return "gt"
	case ">=":
		return "gt_equals"

	// maths
	case "+":
		return "plus"
	case "-":
		return "minus"
	case "*":
		return "multiply"
	case "/":
		return "divide"
	case "%":
		return "modulus"

	// type checks
	case "cons?":
		return "consp"
	case "char?":
		return "charp"
	case "float?":
		return "floatp"
	case "int?":
		return "intp"
	case "lambda?":
		return "lambdap"
	case "nil?":
		return "nilp"
	case "numeric?":
		return "numericp"
	case "str?":
		return "strp"
	}

	// Rewrite some names to avoid errors from nasm.
	// e.g. creating function "foo:bar" would try to
	// define a label "foo:bar:" and that would be rejected.
	specials := []string{":", "-", "!"}
	for _, str := range specials {
		name = strings.ReplaceAll(name, str, "_")
	}

	// other functions just get "fn_" prefix
	if strings.HasPrefix(name, "fn_") {
		return name
	}
	return "fn_" + name
}

// emitExpr emits the code for each of our expression AST types.
func (c *Compiler) emitExpr(e parser.Expr, ev *env.Env) error {
	switch n := e.(type) {

	case *parser.Call:
		if symbol, ok := n.Fn.(*parser.Symbol); ok {

			// is this variadic?
			name := symbol.Name
			v, ok := c.functions[name]

			// Failed to lookup - look for the package-local function
			if !ok {
				nm := c.inPackage + ":" + symbol.Name
				vv, ok2 := c.functions[nm]

				// Okay that worked.  Rename
				if ok2 {
					v = vv
					ok = ok2
					name = nm
				}
			}

			if ok && v.Variadic {

				// Variadic call.
				err := c.emitVariadicCall(name, v.Arguments, n.Args, ev)
				return err
			}

			regs := []string{
				"rdi",
				"rsi",
				"rdx",
				"rcx",
				"r8",
				"r9",
			}

			// Mismatch in argument counts?
			if ok {
				if len(n.Args) != v.Arguments {
					return fmt.Errorf("arity-error: function %s expects %d arguments, %d provided", name, v.Arguments, len(n.Args))
				}
			}

			for _, a := range n.Args {
				err := c.emitExpr(a, ev)
				if err != nil {
					return err
				}
				c.emitln("    push rax")
			}

			for i := len(n.Args) - 1; i >= 0; i-- {
				c.emitln(fmt.Sprintf(
					"    pop %s",
					regs[i],
				))
			}

			// lambda?
			// This covers the case where  a lambda is stored in the
			// environment/symbol table, bound to a variable, such as
			//
			//       (let ((x (lambda (a b) (+ a b))))
			//         (println (x 3 7)))
			//
			if offset, ok := ev.Lookup(name); ok {

				c.emitln(fmt.Sprintf(
					"    mov rax,[rbp-%d]",
					offset,
				))

				// check if it is a lambda
				c.emitln("mov rbx,rax")
				c.emitln("GET_TAG_BITS rbx")
				c.emitln("cmp rbx, TAG_ID_LAMBDA")
				c.emitln("jne type_error")

				// call the lambda
				c.emitln("UNTAG_REG rax")
				c.emitln("mov r15, rax")
				c.emitln("mov rax, [r15]")
				c.emitln("call rax")

				return nil
			}

			// Similar story here - a lambda that is stored in a global
			// variable instead of a local one
			if _, ok := c.globals[name]; ok {

				// get the address
				c.emitln(fmt.Sprintf("    mov rax,[%s]  ; %s", c.addThing("global", name), name))

				// check if it is a lambda
				c.emitln("mov rbx,rax")
				c.emitln("GET_TAG_BITS rbx")
				c.emitln("cmp rbx, TAG_ID_LAMBDA")
				c.emitln("jne type_error")

				// call the lambda
				c.emitln("UNTAG_REG rax")
				c.emitln("mov r15, rax")
				c.emitln("mov rax, [r15]")
				c.emitln("call rax")

				return nil
			}

			// OK then we assume it's a function
			c.emitln("    call " + c.asmName(name))
			return nil
		}

		//
		// This covers an (anonymous) inline lambda which isn't
		// stored in our symbol/environment table such as:
		//  (println ( (lambda (a) (/ 100 a)) 10))
		//
		regs := []string{
			"rdi",
			"rsi",
			"rdx",
			"rcx",
			"r8",
			"r9",
		}

		for _, a := range n.Args {
			err := c.emitExpr(a, ev)
			if err != nil {
				return err
			}

			c.emitln("    push rax")
		}

		for i := len(n.Args) - 1; i >= 0; i-- {
			c.emitln(fmt.Sprintf(
				"    pop %s",
				regs[i],
			))
		}

		// evaluate callable expression
		err := c.emitExpr(n.Fn, ev)
		if err != nil {
			return err
		}

		// check if it is a lambda
		c.emitln("mov rbx,rax")
		c.emitln("GET_TAG_BITS rbx")
		c.emitln("cmp rbx, TAG_ID_LAMBDA")
		c.emitln("jne type_error")

		// call the lambda
		c.emitln("UNTAG_REG rax")
		c.emitln("mov r15, rax")
		c.emitln("mov rax, [r15]")
		c.emitln("call rax")

	case *parser.Cond:

		label := c.label("cond_")

		// There are N test/bodies - compile the comparisons to jump to
		// each body
		for i, cas := range n.Cases {

			err := c.emitExpr(cas.Case, ev)
			if err != nil {
				return err
			}

			c.emitln("    GET_TAG_BITS rax     ; get type bits")
			c.emitln("    cmp rax, TAG_ID_NIL  ; is this a nil?")
			c.emitln(fmt.Sprintf("     jnz %s_case_%d", label, i))
		}

		// No match? Then fall-through to return nil
		c.emitln(label + "_nil:")
		c.emitln("   xor rax, rax")
		c.emitln("   TAG_NIL_REG rax")
		c.emitln(fmt.Sprintf("   jmp %s_end", label))

		// now compile each body - making sure execution jumps to the end
		for i, cas := range n.Cases {

			// case for each one
			c.emitln(fmt.Sprintf("%s_case_%d:", label, i))
			for _, expr := range cas.Exprs {
				err := c.emitExpr(expr, ev)
				if err != nil {
					return err
				}
			}
			c.emitln(fmt.Sprintf("   jmp %s_end", label))
		}

		// define end
		c.emitln(label + "_end:")

	case *parser.Char:
		c.emitln(fmt.Sprintf("    mov rax, %d", n.Value))
		c.emitln("   TAG_CHAR_REG rax")

	case *parser.Do:
		for _, expr := range n.Exprs {
			err := c.emitExpr(expr, ev)
			if err != nil {
				return err
			}
		}

	case *parser.Float:

		// create a label, based on the hash of the content.
		// This has the side-effect of interning.
		lbl := c.addThing("float", n.Value)

		c.floats[lbl] = n.Value

		// load the address of the label and tag.
		// same as our string-handling.
		c.emitln(fmt.Sprintf("    lea rax, %s", lbl))
		c.emitln("    TAG_FLOAT_REG rax")

	case *parser.Int:
		c.emitln(fmt.Sprintf("    mov rax, %d", n.Value))
		c.emitln("   TAG_INTEGER_REG rax")

	case *parser.If:
		elseLbl := c.label("else")
		endLbl := c.label("endif")

		err := c.emitExpr(n.Cond, ev)
		if err != nil {
			return err
		}

		c.emitln("    GET_TAG_BITS rax     ; get type bits")
		c.emitln("    cmp rax, TAG_ID_NIL  ; is this a nil?")
		c.emitln("    jz " + elseLbl)

		err = c.emitExpr(n.Then, ev)
		if err != nil {
			return err
		}

		c.emitln("    jmp " + endLbl)

		c.emitln(elseLbl + ":")

		// else branch is optional
		if n.Else != nil {
			err = c.emitExpr(n.Else, ev)
			if err != nil {
				return err
			}

		}
		c.emitln(endLbl + ":")

	case *parser.Lambda:

		// create a unique name for this lambda
		name := c.asmName(fmt.Sprintf("lambda_%d", c.labelID))
		c.labelID++

		// We don't do analysis for captured variables,
		// we just claim ALL of them.
		n.Captures = ev.Names()

		// Allocate closure:
		//   +0  code pointer
		//   +8  capture #1
		//   +16 capture #2
		//   ...
		size := 8 * (1 + len(n.Captures))

		c.emitln("    mov rax, [heap_ptr]")
		c.emitln(fmt.Sprintf(
			"    add qword [heap_ptr], %d",
			size,
		))
		c.emitln("    mov rbx, rax")

		// store code pointer
		c.emitln(fmt.Sprintf(
			"    mov qword [rbx], %s",
			name,
		))

		for i, cap := range n.Captures {

			if offset, ok := ev.Lookup(cap); ok {
				c.emitln(fmt.Sprintf(
					"    mov rcx,[rbp-%d]",
					offset,
				))
			} else if offset, ok := ev.LookupCapture(cap); ok {
				c.emitln(fmt.Sprintf(
					"    mov rcx,[r15+%d]",
					offset,
				))
			} else {
				panic("capture not found: " + cap)
			}

			c.emitln(fmt.Sprintf(
				"    mov [rbx+%d], rcx",
				8*(i+1),
			))
		}

		// return tagged closure
		c.emitln("    mov rax, rbx")
		c.emitln("    TAG_LAMBDA_REG rax")

		// save away the lambda in the list of lambdas
		n.Name = name
		c.lambdas = append(c.lambdas, n)

	case *parser.Let:
		// create a new child environment
		child := env.New(ev)

		// populate the new environment
		for _, b := range n.Bindings {

			// define the name before we compile
			// the expression.
			offset := child.Define(b.Name)

			// now the expression - but we
			// give that a reference to the
			// child environment, so that
			// references to earlier bindings
			// work as we want.
			//
			// i.e. We want "(let* ..)" rather
			// than "(let ..)"
			err := c.emitExpr(b.Expr, child)
			if err != nil {
				return err
			}

			// and store the result in the
			// binding-reservation.
			c.emitln(fmt.Sprintf(
				"    mov [rbp-%d], rax",
				offset,
			))
		}

		// compile each expression within the body
		for _, expr := range n.Body {
			err := c.emitExpr(expr, child)
			if err != nil {
				return err
			}
		}

	case *parser.Nil:
		c.emitln("    xor rax, rax     ; NIL")
		c.emitln("    TAG_NIL_REG rax  ; Tagged")

	case parser.Package:
		if c.inPackage != "" {
			return fmt.Errorf("nested packages are illegal, in package %s new package %s", c.inPackage, n.Name)
		}

		// store the package
		c.inPackage = n.Name

		// emit all the expressions
		for _, expr := range n.Contents {
			err := c.emitExpr(expr, ev)
			if err != nil {
				return err
			}
		}

		// clear the package
		c.inPackage = ""

	case *parser.String:
		// create a label, based on the hash of the content.
		// This has the side-effect of interning.
		lbl := c.addThing("string", n.Value)

		// save the string, because we're gonna put it into the
		// generated code, later.
		c.strings[lbl] = n.Value

		// load the address of the label and tag.
		// same as our float-handling.
		c.emitln(fmt.Sprintf("    lea rax, %s", lbl))
		c.emitln("    TAG_STRING_REG rax")

	case *parser.Set:

		// If we're in a package the name changes
		if c.inPackage != "" {
			n.Name = c.inPackage + ":" + n.Name
		}

		err := c.emitExpr(n.Expr, ev)
		if err != nil {
			return err
		}

		if offset, ok := ev.Lookup(n.Name); ok {
			c.emitln(fmt.Sprintf(
				"    mov [rbp-%d], rax",
				offset,
			))
			return nil
		}

		if offset, ok := ev.LookupCapture(n.Name); ok {
			c.emitln(fmt.Sprintf(
				"    mov [r15+%d], rax",
				offset,
			))
			return nil
		}

		if global, ok := c.globals[n.Name]; ok {
			if global.Const && global.Init {
				return fmt.Errorf("attempt to modify the global constant variable %s", global.Name)
			}

			c.emitln(fmt.Sprintf("    mov [%s], rax  ; %s", c.addThing("global", global.Name), global.Name))
			return nil
		}

		if c.inPackage != "" {
			nm := c.inPackage + ":" + n.Name
			if global, ok := c.globals[nm]; ok {

				if global.Const && global.Init {
					return fmt.Errorf("attempt to modify the package %s constant variable %s", c.inPackage, global.Name)
				}

				c.emitln(fmt.Sprintf("    mov [%s], rax  ; %s", c.addThing("global", nm), nm))
				return nil
			}
		}

		return fmt.Errorf("unknown variable: %s [package '%s']", n.Name, c.inPackage)

	case *parser.Symbol:

		if offset, ok := ev.Lookup(n.Name); ok {
			c.emitln(fmt.Sprintf(
				"    mov rax, [rbp-%d]",
				offset,
			))
			return nil
		}

		if offset, ok := ev.LookupCapture(n.Name); ok {
			c.emitln(fmt.Sprintf(
				"    mov rax, [r15+%d]",
				offset,
			))
			return nil
		}

		if global, ok := c.globals[n.Name]; ok {
			c.emitln(fmt.Sprintf("    mov rax,[%s]  ; %s", c.addThing("global", global.Name), global.Name))
			return nil
		}

		if c.inPackage != "" {
			nm := c.inPackage + ":" + n.Name
			if _, ok := c.globals[nm]; ok {
				c.emitln(fmt.Sprintf("    mov rax, [%s]  ; package %s %s", c.addThing("global", nm), c.inPackage, n.Name))
				return nil
			}
		}

		return fmt.Errorf("unknown variable: %s [package '%s']", n.Name, c.inPackage)

	case *parser.While:

		// create label for now, and the end
		whileStart := c.label("while_start")
		whileEnd := c.label("while_end")

		// We're at the start, where we loop again
		// to test the condition each time
		c.emitln(whileStart + ":")

		// compile the condition
		err := c.emitExpr(n.Cond, ev)
		if err != nil {
			return err
		}

		// If the condition is "nil" we jump
		// to the end.  Otherwise fall through
		// to run the body..
		c.emitln("    GET_TAG_BITS rax     ; get type bits")
		c.emitln("    cmp rax, TAG_ID_NIL  ; is this a nil?")
		c.emitln("    jz " + whileEnd)

		// assemble the body
		for _, expr := range n.Exprs {
			err := c.emitExpr(expr, ev)
			if err != nil {
				return err
			}
		}

		// loop around again
		c.emitln("    jmp " + whileStart)

		// but mark where the body is over.
		c.emitln(whileEnd + ":")

	default:
		return fmt.Errorf("emitExpr: Unhandled node type:%T value:%V", n, n)
	}
	return nil
}

// emitVariadicCall compiles a call to a function which expects a variable number of arguments,
// what this means is that any arguments which are present are converted into a list and passed
// as a single argument.
func (c *Compiler) emitVariadicCall(name string, expected int, args []parser.Expr, ev *env.Env) error {

	regs := []string{
		"rdi",
		"rsi",
		"rdx",
		"rcx",
		"r8",
		"r9",
	}

	//
	// Fixed arguments.
	//
	fixed := 0
	if expected > 0 {
		fixed = expected - 1
	}

	// evaluate fixed args
	for i := 0; i < fixed; i++ {
		if err := c.emitExpr(args[i], ev); err != nil {
			return err
		}
		c.emitln("    push rax")
	}

	//
	// Build a list for all the additional arguments.
	//

	c.emitln("    xor rax,rax")
	c.emitln("    TAG_NIL_REG rax")

	for i := len(args) - 1; i >= fixed; i-- {

		c.emitln("    push rax")

		if err := c.emitExpr(args[i], ev); err != nil {
			return err
		}

		c.emitln("    mov rdi,rax")
		c.emitln("    pop rsi")
		c.emitln("    call fn_cons")
	}

	// Push resulting list.
	c.emitln("    push rax")

	//
	// Pop registers.
	//
	for i := fixed; i >= 0; i-- {
		c.emitln(fmt.Sprintf("    pop %s", regs[i]))
	}

	c.emitln("    call " + c.asmName(name))
	return nil
}

// emitCallable emits the code for either a defun, or a lambda.
//
// The implementation of these is 100% identical EXCEPT the lambda will prefer to
// use captured variables over local ones.  Those are emitted relative to the
// lambda-base environment address, we store in R15.
//
// The Lambda struct actually embeds a Defun one, with the extra capture fields being
// the only difference.
func (c *Compiler) emitCallable(obj any) error {

	// create new environment
	ev := env.New(nil)

	// Case the incoming object into a Defun,
	// because the Lambda node actually embeds on.
	//
	// We do need to add some lambda-specific generation
	// between the prologue and epilogue, but that's small.
	var d *parser.Defun

	switch c := obj.(type) {
	case parser.Defun:
		d = &c
	case *parser.Lambda:
		d = &c.Defun
	default:
		return fmt.Errorf("invalid type %T", obj)
	}

	//
	// Code that is common, and Defun-related
	//
	nm := c.asmName(d.Name)

	// functions go into their own sections
	c.emitln(fmt.Sprintf("section .text.%s,\"ax\",@progbits", nm))
	c.emitln(nm + ":")

	c.emitln("    push rbp")
	c.emitln("    mov rbp, rsp")
	c.emitln("    sub rsp, 256 ;; guess at space for locals")

	regs := []string{
		"rdi",
		"rsi",
		"rdx",
		"rcx",
		"r8",
		"r9",
	}

	for i, p := range d.Params {

		offset := ev.Define(p)

		c.emitln(fmt.Sprintf(
			"    mov [rbp-%d], %s",
			offset,
			regs[i],
		))
	}

	//
	// Lambdas have this extra bit in the middle to emit
	// the capture magic
	//
	l, ok := obj.(*parser.Lambda)
	if ok {
		// define captured variables, relative to our R15 pointer.
		for _, cap := range l.Captures {
			ev.DefineCapture(cap)
		}
	}

	//
	// Now back to the shared/defun-related epilogue.
	//

	for _, xpr := range d.Exprs {
		err := c.emitExpr(xpr, ev)
		if err != nil {
			return err
		}

	}

	c.emitln("    leave")
	c.emitln("    ret")
	return nil
}
