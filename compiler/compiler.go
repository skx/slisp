// Package compiler is our main workhorse, which creates an assembly
// language version of the given input program and outputs it to STDOUT.
package compiler

import (
	"bytes"
	"crypto/sha1"
	"embed"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"

	"github.com/skx/slisp/env"
	"github.com/skx/slisp/parser"
)

//go:embed template.tmpl
var tmplTxt string

// registerArguments are the registers which are used for passing arguments in the
// Sys V ABI.
var registerArguments = []string{
	"rdi",
	"rsi",
	"rdx",
	"rcx",
	"r8",
	"r9",
}

// labelRemapping contains a lookup table of characters that must be remapped
// when generating NASM labels. We could replace illegal (non alphanumeric)
// characters with just "_", but that would risk collisions if we had functions
// named both "foo?" and "foo!".
var labelRemapping = map[string]string{
	":": "COLON",
	"-": "MINUS",
	"+": "PLUS",
	"*": "STAR",
	"!": "BANG",
	"?": "QUESTION",
	"%": "PERCENT",
	">": "GT",
	"<": "LT",
	"=": "EQUALS",
	"/": "DIVIDE",
}

// FunctionArgs records the arguments which a given defun accepts.
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

	// aliases handles renaming user-visible names to
	// assembly routines
	aliases map[string]string

	// fs is the internal filesystem from which packages are loaded.
	fs embed.FS

	// source stores the program we're parsing.
	source string

	// stdlib is our embedded standard library
	stdlib string

	// text stores the text we emit as we compile various things.
	text bytes.Buffer

	// labelID is used to give unique labels to if/lambda/etc.
	labelID int

	// loaded contains packages we've already loaded.
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
}

// New is our constructor
func New(src string) *Compiler {

	// return a new object, with the source and
	// all internal maps created.
	return &Compiler{
		aliases:   map[string]string{},
		source:    src,
		floats:    map[string]float64{},
		functions: map[string]*FunctionArgs{},
		globals:   map[string]parser.Global{},
		loaded:    map[string]bool{},
		strings:   map[string]string{},
	}
}

// SetStdLib allows embedding the standard library
func (c *Compiler) SetStdLib(s string) {

	var b strings.Builder

	b.WriteString("db ")

	for i, c := range []byte(s) {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "0x%02X", c)
	}

	c.stdlib = b.String() + "\ndb 0x00\n"
}

// LoadPackages will enable loading packages from the specified embedded filesystem.
func (c *Compiler) LoadPackages(fs embed.FS) {
	c.fs = fs
}

// findPackage tries to find the location from which to load .lisp files via "(require foo)".
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

		// Try to load the given content from the embedded filesystem
		data, err := c.fs.ReadFile("packages/" + file)
		if err != nil {

			// Error loading from inline.
			//
			// Load from the filesystem.
			path := ""
			path, err = c.findPackage(file)
			if err != nil {
				return nil, err
			}

			data, err = os.ReadFile(path)
			if err != nil {
				return nil, err
			}
		}

		p := parser.New(string(data))

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

// walkTopLevel is a utility function which makes it possible to iterate over
// top-level objects.
//
// We walk over the top-level functions a lot, to generate symbols, compile
// defuns & etc, so this abstraction helps a lot.
func (c *Compiler) walkTopLevel(
	defs []parser.TopLevel,
	fn func(tl parser.TopLevel) error,
) error {

	for _, tl := range defs {

		if err := fn(tl); err != nil {
			return err
		}
	}

	return nil
}

// Compile creates and returns the assembly language source for the given
// list of functions.
func (c *Compiler) Compile() (string, error) {

	// Create a parser object with our source.
	p := parser.New(c.source)

	// Parse the program into top-level items.
	defs, err := p.Parse()
	if err != nil {
		return "", fmt.Errorf("error parsing program %s", err)
	}

	// Walk over the generated AST and process any (require ..)
	// statements, recursively.
	defs, err = c.expandRequires(defs)
	if err != nil {
		return "", err
	}

	//
	// Walk over the top-level functions and record the
	// names of all global functions.
	//
	// Also record details of all known functions and record the number
	// of arguments they request, and whether the last argument should
	// be treated as variadic.
	//
	err = c.walkTopLevel(defs, func(tl parser.TopLevel) error {

		switch n := tl.(type) {

		case parser.Global:

			name := n.Name
			c.globals[name] = n

		case parser.Defun:

			name := n.Name

			c.functions[name] = &FunctionArgs{
				Arguments: len(n.Params),
				Variadic:  n.Variadic,
			}
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	//
	// Walk over the top-level functions and handle any aliasing updates
	// these will change calls from "(old ..)" to "(new ..)" and ensure
	// the parameters match.
	//
	// This has to happen after functons have been recorded, and parameters
	// recorded.
	//
	err = c.walkTopLevel(defs, func(tl parser.TopLevel) error {

		switch n := tl.(type) {

		case parser.Alias:

			// strip quotes, if present.
			old := strings.Trim(n.Old, "\"")
			new := strings.Trim(n.New, "\"")

			_, ok := c.functions[new]
			if !ok {
				return fmt.Errorf("failed to find details for function %s (%s)", n.New, new)
			}
			c.functions[old] = c.functions[new]
			c.aliases[old] = c.asmName(new)
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	//
	// Create a new environment for the global defun/defvar
	// statements - they can't really use it, but it is required.
	//
	e := env.New(nil)

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

	///
	/// Now we compile
	///

	//
	// Walk over all top-level expressions, and handle the setup of global variables.
	//
	err = c.walkTopLevel(defs, func(tl parser.TopLevel) error {
		g, ok := tl.(parser.Global)
		if !ok {
			return nil
		}

		if err = c.emitExpr(g.Value, e); err != nil {
			return err
		}

		name := g.Name

		x := c.globals[name]
		x.Init = true
		c.globals[name] = x

		c.emitln(fmt.Sprintf(
			"mov [%s], rax",
			c.addThing("global", name),
		))

		return nil
	})
	if err != nil {
		return "", err
	}

	//
	// Compiled code to setup the initial value of
	// each known defvar/defconst.
	//
	// To be inserted into our rendered template shortly.
	//
	initGlobals := getCompiled()

	//
	// Have we seen a "main" function, at the top-level
	// (i.e. outside a package).
	//
	main := false

	//
	// We want to allow later functions to override earlier
	// ones.
	//
	// So we iterate over our functions and save them in
	// a hash - before processing that.
	//
	// This way:
	//   (defun foo () (print "OK"))
	//   (defun foo () (print "Hello, World"))
	//
	// Will mean "(foo) -> Hello, World"
	tmp := make(map[string]parser.Defun)

	//
	// Generate the assembly for each known user-defined
	// function to our internal buffer.
	//
	err = c.walkTopLevel(defs, func(tl parser.TopLevel) error {

		d, ok := tl.(parser.Defun)
		if !ok {
			return nil
		}
		tmp[d.Name] = d
		return nil
	})
	if err != nil {
		return "", err
	}

	//
	// Now emit them for real
	//
	for name, ent := range tmp {
		if name == "main" {
			main = true
		}

		err = c.emitCallable(ent)
		if err != nil {
			return "", err
		}
		c.emitln("")
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
		c.emitln("align 16")
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
		c.emitln("align 16")
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

		// Stdlib embeds our standard library
		StdLib string

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
		StdLib:      c.stdlib,
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

	// Is this an aliased function?  Then return
	// the replacement.
	renamed, ok := c.aliases[name]
	if ok {
		return renamed
	}

	// Remap special characters in a humane way.
	//
	// This allows me to implement the primitive "nth!"
	// as an assembly routine named "fn_nthBANG" without
	// having to call it "fn_bang_" which might collide with
	// "fn_bang_" as generated by "(defun bang? ..)"
	//
	// The more specific replacements are good for avoiding that.
	for key, val := range labelRemapping {
		name = strings.ReplaceAll(name, key, val)
	}

	// But any other non-letter/numeric is just renamed
	tmp := ""
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			tmp += "_"
		} else {
			tmp += string(r)
		}
	}
	name = tmp

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

			if ok && v.Variadic {

				// Variadic call.
				err := c.emitVariadicCall(name, v.Arguments, n.Args, ev)
				return err
			}

			// Mismatch in argument counts?
			if ok {
				if len(n.Args) != v.Arguments {
					return fmt.Errorf("arity-error: function %s expects %d arguments, %d provided", name, v.Arguments, len(n.Args))
				}
			}

			if len(n.Args) >= 5 {
				return fmt.Errorf("%d is more than the maximum number of arguments we support", len(n.Args))
			}

			//
			// Evaluate each argument and stash them on the frame.
			//
			// In the past we pushed to the stack, but that meant that the values
			// were invisible to our GC process and we'd inevitably die with some
			// corruption in the future.
			//
			argTmp := make([]int, len(n.Args))
			for i, a := range n.Args {
				err := c.emitExpr(a, ev)
				if err != nil {
					return err
				}
				argTmp[i] = ev.NewTemp()
				c.emitln(fmt.Sprintf("    mov [rbp-%d], rax", argTmp[i]))
			}

			// Load them up.
			for i := range n.Args {
				c.emitln(fmt.Sprintf(
					"    mov %s, [rbp-%d]",
					registerArguments[i],
					argTmp[i],
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

		if len(n.Args) >= 5 {
			return fmt.Errorf("%d is more than the maximum number of arguments we support", len(n.Args))
		}

		//
		// Here we go again.
		//
		// I don't love the duplication we have here.
		//
		// Stash args on the frame, not on the stack, so they are visible to GC.
		//
		argTmp := make([]int, len(n.Args))
		for i, a := range n.Args {
			err := c.emitExpr(a, ev)
			if err != nil {
				return err
			}

			argTmp[i] = ev.NewTemp()
			c.emitln(fmt.Sprintf("    mov [rbp-%d], rax", argTmp[i]))
		}

		// evaluate callable expression
		err := c.emitExpr(n.Fn, ev)
		if err != nil {
			return err
		}

		// The callable might itself be a heap-allocated (lambda) value,
		// so it also needs to stay in a tracked slot while we load the
		// argument registers below.
		fnTmp := ev.NewTemp()
		c.emitln(fmt.Sprintf("    mov [rbp-%d], rax", fnTmp))

		for i := range n.Args {
			c.emitln(fmt.Sprintf(
				"    mov %s, [rbp-%d]",
				registerArguments[i],
				argTmp[i],
			))
		}
		c.emitln(fmt.Sprintf("    mov rax, [rbp-%d]", fnTmp))

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
		//   +8   n captures
		//   +16  capture #1
		//   +24  capture #2
		//   ...
		size := 8 * (12 + len(n.Captures))

		c.emitln(fmt.Sprintf(
			"     mov rax, %d",
			size,
		))
		c.emitln("    push rbx")
		c.emitln("    mov rbx, TAG_ID_LAMBDA")
		c.emitln("    call alloc")
		c.emitln("    pop rbx")

		c.emitln("    mov rbx, rax")

		// store code pointer
		c.emitln(fmt.Sprintf(
			"    mov qword [rbx], %s",
			name,
		))
		// store N captures
		c.emitln(fmt.Sprintf(
			"    mov qword [rbx+8], %d",
			len(n.Captures),
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
					offset+8, // skip over N captures
				))
			} else {
				panic("capture not found: " + cap)
			}

			c.emitln(fmt.Sprintf(
				"    mov [rbx+%d], rcx",
				8*(i+2),
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
		name := n.Name

		err := c.emitExpr(n.Expr, ev)
		if err != nil {
			return err
		}

		if offset, ok := ev.Lookup(name); ok {
			c.emitln(fmt.Sprintf(
				"    mov [rbp-%d], rax",
				offset,
			))
			return nil
		}

		if offset, ok := ev.LookupCapture(name); ok {
			c.emitln(fmt.Sprintf(
				"    mov [r15+%d], rax",
				offset+8, // skip over N captures
			))
			return nil
		}

		if global, ok := c.globals[name]; ok {
			if global.Const && global.Init {
				return fmt.Errorf("attempt to modify the global constant variable %s", global.Name)
			}

			c.emitln(fmt.Sprintf("    mov [%s], rax  ; %s", c.addThing("global", name), global.Name))
			return nil
		}
		return fmt.Errorf("unknown variable: %s", n.Name)

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
				offset+8, // skip over N captures
			))
			return nil
		}

		if global, ok := c.globals[n.Name]; ok {
			c.emitln(fmt.Sprintf("    mov rax,[%s]  ; %s", c.addThing("global", global.Name), global.Name))
			return nil
		}

		return fmt.Errorf("unknown variable: %s", n.Name)

	case *parser.Unless:
		endLbl := c.label("unless")

		err := c.emitExpr(n.Cond, ev)
		if err != nil {
			return err
		}

		c.emitln("    GET_TAG_BITS rax     ; get type bits")
		c.emitln("    cmp rax, TAG_ID_NIL  ; is this a nil?")
		c.emitln("    jnz " + endLbl)

		// assemble the body
		for _, expr := range n.Exprs {
			err := c.emitExpr(expr, ev)
			if err != nil {
				return err
			}
		}

		c.emitln(endLbl + ":")

	case *parser.When:
		endLbl := c.label("when")

		err := c.emitExpr(n.Cond, ev)
		if err != nil {
			return err
		}

		c.emitln("    GET_TAG_BITS rax     ; get type bits")
		c.emitln("    cmp rax, TAG_ID_NIL  ; is this a nil?")
		c.emitln("    jz " + endLbl)

		// assemble the body
		for _, expr := range n.Exprs {
			err := c.emitExpr(expr, ev)
			if err != nil {
				return err
			}
		}

		c.emitln(endLbl + ":")

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

	//
	// Fixed arguments.
	//
	fixed := 0
	if expected > 0 {
		fixed = expected - 1
	}

	//
	// Evaluate each argument and stash them on the frame.
	//
	// In the past we pushed to the stack, but that meant that the values
	// were invisible to our GC process and we'd inevitably die with some
	// corruption in the future.
	//
	fixedTmp := make([]int, fixed)
	for i := 0; i < fixed; i++ {
		if err := c.emitExpr(args[i], ev); err != nil {
			return err
		}
		fixedTmp[i] = ev.NewTemp()
		c.emitln(fmt.Sprintf("    mov [rbp-%d], rax", fixedTmp[i]))
	}

	//
	// Build a list for all the additional arguments.
	//
	c.emitln("    xor rax,rax")
	c.emitln("    TAG_NIL_REG rax")

	//
	// Now build the list for the variadic arguments, once again
	// these must be stored within the frame via RBP, not the
	// stack otherwise GC will ignore them - which means after
	// GC has finished we'll have bogus values.
	//
	listTmp := ev.NewTemp()
	c.emitln(fmt.Sprintf("    mov [rbp-%d], rax", listTmp))

	for i := len(args) - 1; i >= fixed; i-- {

		if err := c.emitExpr(args[i], ev); err != nil {
			return err
		}

		c.emitln("    mov rdi,rax")
		c.emitln(fmt.Sprintf("    mov rsi, [rbp-%d]", listTmp))
		c.emitln("    call fn_cons")
		c.emitln(fmt.Sprintf("    mov [rbp-%d], rax", listTmp))
	}

	//
	// Load the register values via the frame pointer we setup above.
	//
	for i := 0; i < fixed; i++ {
		c.emitln(fmt.Sprintf("    mov %s, [rbp-%d]", registerArguments[i], fixedTmp[i]))
	}
	c.emitln(fmt.Sprintf("    mov %s, [rbp-%d]", registerArguments[fixed], listTmp))

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

	name := d.Name

	//
	// Code that is common, and Defun-related
	//
	nm := c.asmName(name)

	//
	// Avoid duplication
	//
	_, renamed := c.aliases[name]
	if renamed {
		return nil
	}

	// functions go into their own sections
	c.emitln(fmt.Sprintf("section .text.%s,\"ax\",@progbits", nm))
	c.emitln(nm + ":")

	if len(d.Params) >= 5 {
		return fmt.Errorf("%d is more than the maximum number of arguments we support", len(d.Params))
	}

	// Buffer the function body so we can determine the exact stack frame
	// size (MaxOffset) before emitting the prologue's sub rsp instruction.
	// This avoids the over-allocation that caused stack overflows in deeply
	// recursive functions.
	savedLen := c.text.Len()

	for i, p := range d.Params {

		offset := ev.Define(p)

		c.emitln(fmt.Sprintf(
			"    mov [rbp-%d], %s",
			offset,
			registerArguments[i],
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

	// Extract the buffered body, truncate back to before we started,
	// then emit the prologue with the exact frame size now that we know it.
	locals := ev.MaxOffset()
	bodyText := strings.Clone(c.text.String()[savedLen:])
	c.text.Truncate(savedLen)

	// Frame size = locals (deepest slot offset) rounded up to 16-byte
	// boundary so the stack stays aligned for nested calls.
	frameSize := (locals + 15) &^ 15

	c.emitln("    push rbp")
	c.emitln("    mov rbp, rsp")
	c.emitln(fmt.Sprintf("    push fn_%s_gc", nm))
	c.emitln(fmt.Sprintf("    sub rsp, %d", frameSize))

	// Zero-initialize all local slots so the GC always sees valid tagged
	// values (integer 0) even when sys-gc is called before locals are assigned.
	for off := 16; off <= locals; off += 8 {
		c.emitln(fmt.Sprintf("    mov qword [rbp-%d], 0", off))
	}

	c.text.WriteString(bodyText)

	c.emitln("    leave")
	c.emitln("    ret")

	localBytes := locals - 8
	if localBytes < 0 {
		localBytes = 0
	}
	c.emitln("section .data")
	c.emitln(fmt.Sprintf("fn_%s_gc:", nm))
	c.emitln("dq 0x47430001     ; GC01")
	c.emitln(fmt.Sprintf("dq %d", localBytes))
	c.emitln("section .text")

	return nil
}
