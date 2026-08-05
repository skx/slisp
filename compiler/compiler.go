// Package compiler is our main workhorse, which creates an assembly
// language version of the given input program and outputs it to STDOUT.
package compiler

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
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

// maxMacroDepth guards against a macro which expands into a call to
// itself (directly, or indirectly).
var maxMacroDepth = 500

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

	// macros stores the known "defmacro" definitions.
	macros map[string]parser.Defmacro

	// macroDepth tracks how many macro-expansions are currently nested,
	// this is just to avoid recursion limits.
	macroDepth int
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
		macros:    map[string]parser.Defmacro{},
		loaded:    map[string]bool{},
		strings:   map[string]string{},
	}
}

// SetStdLib allows embedding the standard library
func (c *Compiler) SetStdLib(s string) error {

	// Before we encoded it strip the comments.
	s, err := c.trimComments(s)
	if err != nil {
		return err
	}

	var b strings.Builder

	b.WriteString("db ")

	for i, c := range []byte(s) {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "0x%02X", c)
	}

	c.stdlib = b.String() + "\ndb 0x00\n"
	return nil
}

// trimComments processes the given text, and returns a copy without
// [lisp] comments.
func (c *Compiler) trimComments(str string) (string, error) {
	var out bytes.Buffer

	scanner := bufio.NewScanner(strings.NewReader(str))

	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		// Remove trailing comment.
		if strings.HasPrefix(line, ";") {
			line = strings.TrimPrefix(line, ";")
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		out.WriteString(line)
		out.WriteByte('\n')
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return out.String(), nil
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

		case parser.Defmacro:

			c.macros[n.Name] = n
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
	// Compute the frame size, and the number of root-bytes the GC
	// should scan for the variables setup as globals.  Just like we
	// do for functions.
	//
	initGlobalsLocals := e.MaxOffset()
	initGlobalsFrameSize := (initGlobalsLocals + 15) &^ 15
	initGlobalsRootBytes := initGlobalsLocals - 8
	if initGlobalsRootBytes < 0 {
		initGlobalsRootBytes = 0
	}

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
	tmp := map[string]parser.Defun{}

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
	for i := 0; i < len(c.lambdas); i++ {
		err := c.emitCallable(c.lambdas[i])
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
	// Define a structure to hold static strings,
	// from our string-table
	//
	type String struct {
		Name  string
		Value string
	}

	stringLiterals := []String{}
	for id, str := range c.strings {
		stringLiterals = append(stringLiterals,
			String{
				Name:  id,
				Value: strings.ReplaceAll(str, "`", "\\`"),
			})
	}

	//
	// Define a structure to hold static floats,
	// from our float-table
	//
	type Float struct {
		Name  string
		Value string
	}

	floatLiterals := []Float{}
	for id, str := range c.floats {
		floatLiterals = append(floatLiterals, Float{
			Name:  id,
			Value: fmt.Sprintf("%f", str)})
	}

	//
	// Define a structure to hold embedded assets
	//
	// This is NOT used for our stdlib, but it is used for
	// our embedded package-files, which come from packages/
	//
	type Asset struct {
		Name string

		Data string
	}

	//
	// Generate assets from the files we embed in packages/
	//
	assets := []Asset{}

	err = fs.WalkDir(c.fs, ".", func(path string, d fs.DirEntry, err error) error {
		if d.Type().IsRegular() {
			data, err := c.fs.ReadFile(path)
			if err != nil {
				return err
			}

			// Trim the data
			txt, err := c.trimComments(string(data))
			if err != nil {
				return err

			}

			// strip directory
			name := filepath.Base(path)
			// strip suffix
			name = strings.TrimSuffix(name, filepath.Ext(name))

			var b strings.Builder
			for i, c := range txt {
				if i > 0 {
					b.WriteString(", ")
				}
				fmt.Fprintf(&b, "0x%02X", c)
			}

			assets = append(assets, Asset{
				Name: name,
				Data: b.String(),
			})
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	//
	// Sort assets by name: in reverse because "packages" will
	// return them in this order.
	//
	sort.Slice(assets, func(i, j int) bool {
		return assets[i].Name > assets[j].Name
	})

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
		// Static assets
		Assets []Asset

		// Count of assets
		AssetCount int

		// The defintions of defun's we've seen.
		Defuns string

		// Floattable holds our floating point constants
		FloatTable []Float

		// Lambdas has all the lambda expressions we've seen.
		Lambdas string

		// InitGlobals is the thing that loads global variables
		InitGlobals string

		// InitGlobalsFrameSize is the stack frame size required
		// for the global variables, setup ahead of main.
		InitGlobalsFrameSize int

		// InitGlobalsRootBytes similar story, bytes vs. size.
		//
		// To be honest this stuff needs to be reworked for all the
		// places we care about stack size/bytes. (i.e. emitCallable).
		InitGlobalsRootBytes int

		// Globals has global variables
		Globals []string

		// StringTable contains the strings we've seen.
		StringTable []String

		// Stdlib embeds our standard library
		StdLib string
	}

	//
	// Create an instance of that internal structure, which we
	// can then pass to the template processor to fill out into
	// the template appropriately.
	//
	x := &Generated{
		AssetCount:           len(assets),
		Assets:               assets,
		Defuns:               defuns,
		FloatTable:           floatLiterals,
		Globals:              globals,
		InitGlobals:          initGlobals,
		InitGlobalsFrameSize: initGlobalsFrameSize,
		InitGlobalsRootBytes: initGlobalsRootBytes,
		Lambdas:              lambdas,
		StdLib:               c.stdlib,
		StringTable:          stringLiterals,
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
		// Is this a function call?
		if symbol, ok := n.Fn.(*parser.Symbol); ok {

			// expand macro, if necessary
			if macro, ok := c.macros[symbol.Name]; ok {

				if c.macroDepth >= maxMacroDepth {
					return fmt.Errorf("macro %s: expansion nested too deeply (possible infinite recursion)", symbol.Name)
				}

				c.macroDepth++
				expanded, err := c.expandMacro(symbol.Name, macro, n.Args)
				if err != nil {
					c.macroDepth--
					return err
				}

				err = c.emitExpr(expanded, ev)
				c.macroDepth--
				return err
			}

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

			if len(n.Args) > len(registerArguments) {
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

			//
			// The lambda might be stored in a captured-variable,
			// or closure, and that's valid too.
			//
			// We need this for the Z-combinator..
			//
			if offset, ok := ev.LookupCapture(name); ok {

				c.emitln(fmt.Sprintf(
					"    mov rax,[r15+%d]",
					offset+8,
				))

				c.emitln("mov rbx,rax")
				c.emitln("GET_TAG_BITS rbx")
				c.emitln("cmp rbx, TAG_ID_LAMBDA")
				c.emitln("jne type_error")

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

		if len(n.Args) > len(registerArguments) {
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

	case *parser.Int:
		c.emitln(fmt.Sprintf("    mov rax, %d", n.Value))
		c.emitln("   TAG_INTEGER_REG rax")

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

	case *parser.List:
		// Build a list - this is as a result of a macro.
		return c.emitExpr(c.evalToList(n.Elems), ev)

	case *parser.Nil:
		c.emitln("    xor rax, rax     ; NIL")
		c.emitln("    TAG_NIL_REG rax  ; Tagged")

	case *parser.Quasiquote:
		expr, err := c.quoteToExpr(n.Expr, true)
		if err != nil {
			return err
		}
		return c.emitExpr(expr, ev)

	case *parser.Quote:
		expr, err := c.quoteToExpr(n.Expr, false)
		if err != nil {
			return err
		}
		return c.emitExpr(expr, ev)

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

	case *parser.Unquote:
		return fmt.Errorf("unquote (,) may only appear within a quasiquote")

	case *parser.UnquoteSplicing:
		return fmt.Errorf("unquote-splicing (,@) may only appear as a list-element within a quasiquote")

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

// evalToList creates a list by calling (cons) appropriately.
func (c *Compiler) evalToList(elems []parser.Expr) parser.Expr {
	var acc parser.Expr = &parser.Nil{}

	for i := len(elems) - 1; i >= 0; i-- {
		acc = &parser.Call{
			Fn:   &parser.Symbol{Name: "cons"},
			Args: []parser.Expr{elems[i], acc},
		}
	}

	return acc
}

// quoteToExpr converts the body of a Quote/Quasiquote expression
// into an ordinary expression which constructs the equivalent
// literal value at runtime.
func (c *Compiler) quoteToExpr(e parser.Expr, quasi bool) (parser.Expr, error) {

	asData := func(name string, inner parser.Expr) (parser.Expr, error) {
		quoted, err := c.quoteToExpr(inner, false)
		if err != nil {
			return nil, err
		}
		return c.buildQuotedList([]parser.Expr{&parser.Symbol{Name: name}, quoted}, false)
	}

	switch n := e.(type) {

	case *parser.Call:
		elems := append([]parser.Expr{n.Fn}, n.Args...)
		return c.buildQuotedList(elems, quasi)

	case *parser.Char:
		// self-evaluating literal.
		return n, nil

	case *parser.Do:
		elems := append([]parser.Expr{&parser.Symbol{Name: "do"}}, n.Exprs...)
		return c.buildQuotedList(elems, quasi)

	case *parser.Float:
		// self-evaluating literal.
		return n, nil

	case *parser.If:
		elems := []parser.Expr{&parser.Symbol{Name: "if"}, n.Cond, n.Then}
		if n.Else != nil {
			elems = append(elems, n.Else)
		}
		return c.buildQuotedList(elems, quasi)

	case *parser.Int:
		// self-evaluating literal.
		return n, nil

	case *parser.Lambda:
		params := make([]parser.Expr, len(n.Params))
		for i, p := range n.Params {
			name := p
			if n.Variadic && i == len(n.Params)-1 {
				name = "&" + name
			}
			params[i] = &parser.Symbol{Name: name}
		}
		elems := append([]parser.Expr{
			&parser.Symbol{Name: "lambda"},
			&parser.List{Elems: params},
		}, n.Exprs...)
		return c.buildQuotedList(elems, quasi)

	case *parser.Let:
		binds := make([]parser.Expr, len(n.Bindings))
		for i, b := range n.Bindings {
			binds[i] = &parser.List{Elems: []parser.Expr{&parser.Symbol{Name: b.Name}, b.Expr}}
		}
		elems := append([]parser.Expr{
			&parser.Symbol{Name: "let"},
			&parser.List{Elems: binds},
		}, n.Body...)
		return c.buildQuotedList(elems, quasi)

	case *parser.List:
		return c.buildQuotedList(n.Elems, quasi)

	case *parser.Nil:
		// self-evaluating literal.
		return n, nil

	case *parser.Quote:
		return asData("quote", n.Expr)

	case *parser.Quasiquote:
		return asData("quasiquote", n.Expr)

	case *parser.Set:
		elems := []parser.Expr{&parser.Symbol{Name: "set!"}, &parser.Symbol{Name: n.Name}, n.Expr}
		return c.buildQuotedList(elems, quasi)

	case *parser.String:
		// self-evaluating literal.
		return n, nil

	case *parser.Symbol:
		// conver to a string, just like we do for :foo.
		return &parser.String{Value: n.Name}, nil

	case *parser.Unquote:
		if quasi {
			return n.Expr, nil
		}
		return asData("unquote", n.Expr)

	case *parser.UnquoteSplicing:
		if quasi {
			return nil, fmt.Errorf("unquote-splicing (,@) may only appear as a list-element within a quasiquote")
		}
		return asData("unquote-splicing", n.Expr)

	case *parser.While:
		elems := append([]parser.Expr{&parser.Symbol{Name: "while"}, n.Cond}, n.Exprs...)
		return c.buildQuotedList(elems, quasi)

	default:
		return nil, fmt.Errorf("quote: cannot quote expression of type %T", e)
	}
}

// buildQuotedList converts a list of expressions into an expression which,
// when compiled, builds the equivalent runtime list, via nested "cons" calls.
func (c *Compiler) buildQuotedList(elems []parser.Expr, quasi bool) (parser.Expr, error) {

	var acc parser.Expr = &parser.Nil{}

	for i := len(elems) - 1; i >= 0; i-- {

		if spl, ok := elems[i].(*parser.UnquoteSplicing); ok && quasi {
			acc = &parser.Call{
				Fn:   &parser.Symbol{Name: "append"},
				Args: []parser.Expr{spl.Expr, acc},
			}
			continue
		}

		converted, err := c.quoteToExpr(elems[i], quasi)
		if err != nil {
			return nil, err
		}

		acc = &parser.Call{
			Fn:   &parser.Symbol{Name: "cons"},
			Args: []parser.Expr{converted, acc},
		}
	}

	return acc, nil
}

// expandMacro expands a single call to the given macro, with the given
// (literal, unevaluated) argument expressions, and returns the resulting
// expression - ready to be compiled (or, if it is itself a macro-call,
// expanded further).
func (c *Compiler) expandMacro(name string, macro parser.Defmacro, args []parser.Expr) (parser.Expr, error) {

	fixed := len(macro.Params)
	if macro.Variadic {
		fixed--
	}

	if macro.Variadic {
		if len(args) < fixed {
			return nil, fmt.Errorf("macro %s expects at least %d argument(s), %d provided", name, fixed, len(args))
		}
	} else if len(args) != fixed {
		return nil, fmt.Errorf("macro %s expects %d argument(s), %d provided", name, fixed, len(args))
	}

	bindings := map[string]parser.Expr{}
	for i := 0; i < fixed; i++ {
		bindings[macro.Params[i]] = args[i]
	}
	if macro.Variadic {
		bindings[macro.Params[fixed]] = &parser.List{Elems: append([]parser.Expr{}, args[fixed:]...)}
	}

	var result parser.Expr = &parser.Nil{}
	for _, expr := range macro.Exprs {
		var err error
		result, err = c.evalMacroExpr(expr, bindings)
		if err != nil {
			return nil, fmt.Errorf("error expanding macro %s: %w", name, err)
		}
	}

	return result, nil
}

// Only a restricted subset of expressions is supported.
//
// Arbitrary compile-time computation (e.g. calling "+"
// directly on a macro parameter) is not supported.  You must use
// a quasiquote template to build the code you want to run instead.
func (c *Compiler) evalMacroExpr(e parser.Expr, bindings map[string]parser.Expr) (parser.Expr, error) {
	switch n := e.(type) {

	case *parser.Call:
		return c.evalMacroCall(n, bindings)

	case *parser.Char:
		return n, nil

	case *parser.Float:
		return n, nil

	case *parser.If:
		// A compile-time conditional: only the taken branch is ever
		// evaluated, which is what lets a macro recurse over a
		// variadic parameter until it runs out of arguments.
		cond, err := c.evalMacroExpr(n.Cond, bindings)
		if err != nil {
			return nil, err
		}
		if isMacroTruthy(cond) {
			return c.evalMacroExpr(n.Then, bindings)
		}
		if n.Else == nil {
			return &parser.Nil{}, nil
		}
		return c.evalMacroExpr(n.Else, bindings)

	case *parser.Int:
		return n, nil

	case *parser.Nil:
		return n, nil

	case *parser.Quasiquote:
		return c.evalQuasiquote(n.Expr, bindings, true)

	case *parser.Quote:
		// Quote is always fully literal: no substitution happens,
		// even if it happens to contain what looks like a bound
		// parameter name.
		return n.Expr, nil

	case *parser.String:
		return n, nil

	case *parser.Symbol:
		if bound, ok := bindings[n.Name]; ok {
			return bound, nil
		}
		return nil, fmt.Errorf("unbound symbol %q in macro body", n.Name)

	default:
		return nil, fmt.Errorf("unsupported expression of type %T in macro body: macros may only use bound parameters, literals, if, quote/quasiquote templates, and car/cdr/nil?", e)
	}
}

// isMacroTruthy reports whether a compile-time macro value should be
// treated as "true" by a compile-time "if" - matching the runtime rule
// that only nil (or the empty list) is false.
func isMacroTruthy(e parser.Expr) bool {
	switch n := e.(type) {
	case *parser.Nil:
		return false
	case *parser.List:
		return len(n.Elems) != 0
	default:
		return true
	}
}

// evalMacroCall evaluates a call appearing (outside of any
// quote/quasiquote) within a macro body.  Only "car", "cdr" and "nil?"
// are supported: enough structural list-decomposition for a macro to
// recurse over a variadic parameter, one element at a time.
func (c *Compiler) evalMacroCall(n *parser.Call, bindings map[string]parser.Expr) (parser.Expr, error) {

	sym, ok := n.Fn.(*parser.Symbol)
	if !ok {
		return nil, fmt.Errorf("unsupported call in macro body: the callable must be a bare symbol (car/cdr/nil?)")
	}

	switch sym.Name {
	case "car", "cdr", "nil?":
		// handled below
	default:
		return nil, fmt.Errorf("unsupported function %q called in macro body: only car/cdr/nil? are supported outside of quote/quasiquote templates", sym.Name)
	}

	if len(n.Args) != 1 {
		return nil, fmt.Errorf("%s expects exactly one argument in a macro body, got %d", sym.Name, len(n.Args))
	}

	val, err := c.evalMacroExpr(n.Args[0], bindings)
	if err != nil {
		return nil, err
	}

	switch sym.Name {
	case "nil?":
		elems, convErr := c.asExprList(val)
		if convErr == nil && len(elems) == 0 {
			return &parser.Int{Value: 1}, nil
		}
		return &parser.Nil{}, nil

	case "car":
		elems, convErr := c.asExprList(val)
		if convErr != nil {
			return nil, fmt.Errorf("car: %s", convErr)
		}
		if len(elems) == 0 {
			return nil, fmt.Errorf("car: cannot take the first element of an empty list")
		}
		return elems[0], nil

	case "cdr":
		elems, convErr := c.asExprList(val)
		if convErr != nil {
			return nil, fmt.Errorf("cdr: %s", convErr)
		}
		if len(elems) == 0 {
			return &parser.List{}, nil
		}
		return &parser.List{Elems: elems[1:]}, nil
	}

	return nil, fmt.Errorf("unsupported function %q called in macro body: only car/cdr/nil? are supported outside of quote/quasiquote templates", sym.Name)

}

// evalQuasiquote resolves a quasiquote template used within a macro
// body, substituting any "live" Unquote/UnquoteSplicing holes with the
// bound macro-argument expressions, and returns the result.
//
// Unlike quoteToExpr (used to compile a Quasiquote appearing in regular,
// non-macro, code, which always builds a runtime list-value) this
// preserves the *shape* of the template exactly: an "if" in the
// template stays an *parser.If, ready to compile directly as real code,
// rather than being turned into list-data describing an if-expression.
func (c *Compiler) evalQuasiquote(e parser.Expr, bindings map[string]parser.Expr, quasi bool) (parser.Expr, error) {

	switch n := e.(type) {

	case *parser.Unquote:
		if quasi {
			return c.evalMacroExpr(n.Expr, bindings)
		}
		inner, err := c.evalQuasiquote(n.Expr, bindings, false)
		if err != nil {
			return nil, err
		}
		return &parser.Unquote{Expr: inner}, nil

	case *parser.UnquoteSplicing:
		if quasi {
			return nil, fmt.Errorf("unquote-splicing (,@) may only appear as a list-element within a quasiquote")
		}
		inner, err := c.evalQuasiquote(n.Expr, bindings, false)
		if err != nil {
			return nil, err
		}
		return &parser.UnquoteSplicing{Expr: inner}, nil

	case *parser.Quote:
		// A nested quote is fully inert: it passes through
		// untouched, exactly like at the top of evalMacroExpr.
		return n, nil

	case *parser.Quasiquote:
		inner, err := c.evalQuasiquote(n.Expr, bindings, false)
		if err != nil {
			return nil, err
		}
		return &parser.Quasiquote{Expr: inner}, nil

	case *parser.Symbol, *parser.Int, *parser.Float, *parser.String, *parser.Char, *parser.Nil:
		// Literal data within the template - untouched.
		return n, nil

	case *parser.List:
		elems, err := c.evalQuasiquoteList(n.Elems, bindings, quasi)
		if err != nil {
			return nil, err
		}
		return &parser.List{Elems: elems}, nil

	case *parser.Call:
		fn, err := c.evalQuasiquote(n.Fn, bindings, quasi)
		if err != nil {
			return nil, err
		}
		args, err := c.evalQuasiquoteList(n.Args, bindings, quasi)
		if err != nil {
			return nil, err
		}
		return &parser.Call{Fn: fn, Args: args}, nil

	case *parser.If:
		cond, err := c.evalQuasiquote(n.Cond, bindings, quasi)
		if err != nil {
			return nil, err
		}
		then, err := c.evalQuasiquote(n.Then, bindings, quasi)
		if err != nil {
			return nil, err
		}
		var els parser.Expr
		if n.Else != nil {
			els, err = c.evalQuasiquote(n.Else, bindings, quasi)
			if err != nil {
				return nil, err
			}
		}
		return &parser.If{Cond: cond, Then: then, Else: els}, nil

	case *parser.Set:
		expr, err := c.evalQuasiquote(n.Expr, bindings, quasi)
		if err != nil {
			return nil, err
		}
		return &parser.Set{Name: n.Name, Expr: expr}, nil

	case *parser.Do:
		exprs, err := c.evalQuasiquoteList(n.Exprs, bindings, quasi)
		if err != nil {
			return nil, err
		}
		return &parser.Do{Exprs: exprs}, nil

	case *parser.While:
		cond, err := c.evalQuasiquote(n.Cond, bindings, quasi)
		if err != nil {
			return nil, err
		}
		exprs, err := c.evalQuasiquoteList(n.Exprs, bindings, quasi)
		if err != nil {
			return nil, err
		}
		return &parser.While{Cond: cond, Exprs: exprs}, nil

	case *parser.Let:
		binds := make([]parser.Binding, len(n.Bindings))
		for i, b := range n.Bindings {
			expr, err := c.evalQuasiquote(b.Expr, bindings, quasi)
			if err != nil {
				return nil, err
			}
			binds[i] = parser.Binding{Name: b.Name, Expr: expr}
		}
		body, err := c.evalQuasiquoteList(n.Body, bindings, quasi)
		if err != nil {
			return nil, err
		}
		return &parser.Let{Bindings: binds, Body: body}, nil

	case *parser.Lambda:
		exprs, err := c.evalQuasiquoteList(n.Exprs, bindings, quasi)
		if err != nil {
			return nil, err
		}
		return &parser.Lambda{Defun: parser.Defun{
			Name:     n.Name,
			Params:   n.Params,
			Variadic: n.Variadic,
			Exprs:    exprs,
		}}, nil

	default:
		return nil, fmt.Errorf("quasiquote: cannot process expression of type %T", e)
	}
}

// evalQuasiquoteList processes each element of a quasiquote's list
// content, splicing in the elements of any UnquoteSplicing it finds -
// but only while quasi is true, i.e. we're not inside a further, inert,
// nested quote/quasiquote.
func (c *Compiler) evalQuasiquoteList(elems []parser.Expr, bindings map[string]parser.Expr, quasi bool) ([]parser.Expr, error) {

	var out []parser.Expr

	for _, el := range elems {

		if spl, ok := el.(*parser.UnquoteSplicing); ok && quasi {
			val, err := c.evalMacroExpr(spl.Expr, bindings)
			if err != nil {
				return nil, err
			}

			splElems, err := c.asExprList(val)
			if err != nil {
				return nil, err
			}
			out = append(out, splElems...)
			continue
		}

		expr, err := c.evalQuasiquote(el, bindings, quasi)
		if err != nil {
			return nil, err
		}
		out = append(out, expr)
	}

	return out, nil
}

// asExprList converts an expression which represents a list - a List, a
// Call (reused, elsewhere, to represent a generic head+args list), or
// Nil (the empty list) - into a plain slice of its elements.  It's used
// to splice the result of an UnquoteSplicing (",@") into a surrounding
// list.
func (c *Compiler) asExprList(e parser.Expr) ([]parser.Expr, error) {
	switch n := e.(type) {
	case *parser.List:
		return n.Elems, nil
	case *parser.Nil:
		return nil, nil
	case *parser.Call:
		return append([]parser.Expr{n.Fn}, n.Args...), nil
	default:
		return nil, fmt.Errorf("unquote-splicing (,@) requires a list, got %T", e)
	}
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
		c.emitln("    call fn_sys_cons_NOGC ; No GC during this operation")
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

	// Cast the incoming object into a Defun, because the
	// Lambda node actually embeds one and they are largely identical.
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

	if len(d.Params) > len(registerArguments) {
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
