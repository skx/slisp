// Package compiler is our main workhorse, which creates an assembly
// language version of the given input program and outputs it to STDOUT.
package compiler

import (
	"bytes"
	"crypto/sha1"
	_ "embed"
	"encoding/hex"
	"fmt"
	"strings"
	"text/template"

	"github.com/skx/slisp/env"
	"github.com/skx/slisp/parser"
)

//go:embed template.tmpl
var tmplTxt string

// Compiler holds our state
type Compiler struct {

	// source stores the program we're parsing.
	source string

	// text stores the text we emit as we compile various things.
	text strings.Builder

	// labelID is used to give unique labels to if/lambda/etc.
	labelID int

	// strings holds the strings we've encountered, indexed
	// by their SHA1 sum as ID.  This is how we intern.
	strings map[string]string

	// floats holds the literal floating point numbers  we've encountered,
	// indexed by their SHA1 sum as ID.  This is how we intern.
	floats map[string]float64

	// lambdas holds the lambdas we've encountered.
	lambdas []*parser.Lambda
}

// New is our constructor
func New(src string) *Compiler {

	return &Compiler{source: src}
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

	// Ensure our tables are pristine
	c.strings = map[string]string{}
	c.floats = map[string]float64{}

	defuns := ""
	main := false

	// Generate the user-defined functions to our internal buffer.
	for _, d := range defs {
		if d.Name == "main" {
			main = true
		}
		err = c.emitCallable(d)
		if err != nil {
			return "", err
		}
		c.emitln("")
	}

	if !main {
		return "", fmt.Errorf("There is no entry-point defined; we need a defun named 'main'")
	}

	// Get them, and clear the buffer.
	defuns = c.text.String()
	c.text.Reset()

	// Now user-defined lambdas
	lambdas := ""
	for _, l := range c.lambdas {
		err = c.emitCallable(l)
		if err != nil {
			return "", err
		}

		c.emitln("")
	}
	lambdas = c.text.String()
	c.text.Reset()

	// Then the string-table for user-defined strings
	stringTable := ""
	c.emitln("section .data")
	for id, str := range c.strings {
		c.emitln("align 8")
		c.emitln(id + ":")

		// escape the "`" which are wrapped around the string.
		str = strings.ReplaceAll(str, "`", "\\`")

		c.emitln(fmt.Sprintf("     db `%s`, 0", str))
	}
	stringTable = c.text.String()
	c.text.Reset()

	// Then the literal user-defined floats
	floatTable := ""
	c.emitln("section .data")
	for id, str := range c.floats {
		c.emitln("align 8")
		c.emitln(id + ":")
		c.emitln(fmt.Sprintf("     dq %f", str))
	}
	floatTable = c.text.String()
	c.text.Reset()

	// Define a simple structure we can pass to the text/template
	// file we render for our output
	type Generated struct {
		// The defintions of defun's we've seen.
		Defuns string

		// Lambdas has all the lambda expressions we've seen.
		Lambdas string

		// StringTable contains the strings we've seen.
		StringTable string

		// FloatTable contains the floating point literals we've seen.
		FloatTable string
	}

	// Create an instance to populate the template with
	x := &Generated{
		Defuns:      defuns,
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

// addFloat creates a unique label for our floats,
// based on the SHA1-hash.  Interning them.
func (c *Compiler) addFloat(f float64) string {
	hasher := sha1.New()
	hasher.Write([]byte(fmt.Sprintf("%f", f)))
	sha := hex.EncodeToString(hasher.Sum(nil))
	id := fmt.Sprintf("float_%s", sha)
	return id
}

// addString creates a unique label for our strings,
// based on the SHA1-hash.  Interning them.
func (c *Compiler) addString(str string) string {
	hasher := sha1.New()
	hasher.Write([]byte(str))
	sha := hex.EncodeToString(hasher.Sum(nil))
	id := fmt.Sprintf("str_%s", sha)
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
	case "str?":
		return "strp"
	}

	// allow "-" by rewriting it to _.
	name = strings.ReplaceAll(name, "-", "_")

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

			// lambda?
			if offset, ok := ev.Lookup(symbol.Name); ok {

				c.emitln(fmt.Sprintf(
					"    mov rax,[rbp-%d]",
					offset,
				))

				// call lambda
				c.emitln("UNTAG_REG rax")
				c.emitln("mov r15, rax")
				c.emitln("mov rax, [r15]")
				c.emitln("call rax")

				return nil
			} else {
				// defun
				c.emitln("    call " + c.asmName(symbol.Name))
				return nil
			}
		}

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

		// call lambda
		c.emitln("UNTAG_REG rax")
		c.emitln("mov r15, rax")
		c.emitln("mov rax, [r15]")
		c.emitln("call rax")

	case *parser.Cond:

		label := c.label("case_")

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
		lbl := c.addFloat(n.Value)

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

	case *parser.String:
		// create a label, based on the hash of the content.
		// This has the side-effect of interning.
		lbl := c.addString(n.Value)

		// save the string, because we're gonna put it into the
		// generated code, later.
		c.strings[lbl] = n.Value

		// load the address of the label and tag.
		// same as our float-handling.
		c.emitln(fmt.Sprintf("    lea rax, %s", lbl))
		c.emitln("    TAG_STRING_REG rax")

	case *parser.Set:

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
				offset,
			))
			return nil
		}
		return fmt.Errorf("unknown variable: %s", n.Name)

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
	case *parser.Defun:
		d = c
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
