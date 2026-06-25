// package compiler is our main workhorse, which creates an assembly
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
	// text stores the text we emit as we compile
	text strings.Builder

	// labelID is used to give unique labels to if/lambda/etc
	labelID int

	// strings holds the strings we've encountered, indexed
	// by their SHA1 sum as ID.
	strings map[string]string

	// lambdas holds the lambdas we've encountered and we need
	// to emit those later too.
	lambdas []*parser.Lambda
}

// New is our constructor
func New() *Compiler {
	return &Compiler{}
}

// addString creates a unique label for our strings,
// based on the SHA1-hash.  Interning them.
func (g *Compiler) addString(str string) string {
	hasher := sha1.New()
	hasher.Write([]byte(str))
	sha := hex.EncodeToString(hasher.Sum(nil))
	id := fmt.Sprintf("str_%s", sha)
	return id
}

// label generates a new label, with the given prefix.
func (g *Compiler) label(prefix string) string {
	s := fmt.Sprintf("%s_%d", prefix, g.labelID)
	g.labelID++
	return s
}

// emitln writes a line of assembly/source into our internal buffer.
func (g *Compiler) emitln(s string) {
	g.text.WriteString(s)
	g.text.WriteString("\n")
}

// asmName converts the given label into something nasm will
// accept.  It doesn't like special characters inside label names.
func (g *Compiler) asmName(name string) string {
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
		return "integer_plus"
	case "-":
		return "integer_minus"
	case "*":
		return "integer_multiply"
	case "/":
		return "integer_divide"
	case "%":
		return "integer_modulus"

	// type checks
	case "cons?":
		return "consp"
	case "char?":
		return "charp"
	case "int?":
		return "intp"
	case "lambda?":
		return "lambdap"
	case "nil?":
		return "nilp"
	case "set!":
		return "set"
	case "str?":
		return "strp"
	}

	// other functions just get "fn_" prefix
	return "fn_" + name
}

// emitExpr emits the code for each of our expression AST types.
func (g *Compiler) emitExpr(e parser.Expr, ev *env.Env) {
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
				g.emitExpr(a, ev)
				g.emitln("    push rax")
			}

			for i := len(n.Args) - 1; i >= 0; i-- {
				g.emitln(fmt.Sprintf(
					"    pop %s",
					regs[i],
				))
			}

			// lambda?
			if offset, ok := ev.Lookup(symbol.Name); ok {

				g.emitln(fmt.Sprintf(
					"    mov rax,[rbp-%d]",
					offset,
				))

				// call lambda
				g.emitln("UNTAG_REG rax")
				g.emitln("mov r15, rax")
				g.emitln("mov rax, [r15]")
				g.emitln("call rax")

				return
			} else {
				// defun
				g.emitln("    call " + g.asmName(symbol.Name))
				return
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
			g.emitExpr(a, ev)
			g.emitln("    push rax")
		}

		for i := len(n.Args) - 1; i >= 0; i-- {
			g.emitln(fmt.Sprintf(
				"    pop %s",
				regs[i],
			))
		}

		// evaluate callable expression
		g.emitExpr(n.Fn, ev)

		// call lambda
		g.emitln("UNTAG_REG rax")
		g.emitln("mov r15, rax")
		g.emitln("mov rax, [r15]")
		g.emitln("call rax")

	case *parser.Char:
		g.emitln(fmt.Sprintf("    mov rax, %d", n.Value))
		g.emitln("   TAG_CHAR_REG rax")

	case *parser.Do:
		for _, expr := range n.Exprs {
			g.emitExpr(expr, ev)
		}

	case *parser.Int:
		g.emitln(fmt.Sprintf("    mov rax, %d", n.Value))
		g.emitln("   TAG_INTEGER_REG rax")

	case *parser.If:
		elseLbl := g.label("else")
		endLbl := g.label("endif")

		g.emitExpr(n.Cond, ev)

		g.emitln("    GET_TAG_BITS rax     ; get type bits")
		g.emitln("    cmp rax, TAG_ID_NIL  ; is this a nil?")
		g.emitln("    jz " + elseLbl)

		g.emitExpr(n.Then, ev)

		g.emitln("    jmp " + endLbl)

		g.emitln(elseLbl + ":")

		// else branch is optional
		if n.Else != nil {
			g.emitExpr(n.Else, ev)
		}
		g.emitln(endLbl + ":")

	case *parser.Lambda:

		// create a unique name for this lambda
		name := fmt.Sprintf("lambda_%d", g.labelID)
		g.labelID++

		// We don't do analysis for captured variables,
		// we just claim ALL of them.
		n.Captures = ev.Names()

		// Allocate closure:
		//   +0  code pointer
		//   +8  capture #1
		//   +16 capture #2
		//   ...
		size := 8 * (1 + len(n.Captures))

		g.emitln("    mov rax, [heap_ptr]")
		g.emitln(fmt.Sprintf(
			"    add qword [heap_ptr], %d",
			size,
		))
		g.emitln("    mov rbx, rax")

		// store code pointer
		g.emitln(fmt.Sprintf(
			"    mov qword [rbx], %s",
			name,
		))

		// copy captures
		for i, cap := range n.Captures {

			offset, ok := ev.Lookup(cap)
			if !ok {
				panic("capture not found: " + cap)
			}

			g.emitln(fmt.Sprintf(
				"    mov rcx, [rbp-%d]",
				offset,
			))

			g.emitln(fmt.Sprintf(
				"    mov [rbx+%d], rcx",
				8*(i+1),
			))
		}

		// return tagged closure
		g.emitln("    mov rax, rbx")
		g.emitln("    TAG_LAMBDA_REG rax")

		// save away the lambda in the list of lambdas
		n.Name = name
		g.lambdas = append(g.lambdas, n)

	case *parser.Let:
		// create a new child environment
		child := env.New(ev)

		nextSlot := child.CountLocals()

		// populate the new environment
		for _, b := range n.Bindings {

			g.emitExpr(b.Expr, ev)

			offset := (nextSlot + 1) * 8

			child.Define(b.Name, offset)

			g.emitln(fmt.Sprintf(
				"    mov [rbp-%d], rax",
				offset,
			))

			nextSlot++
		}

		// compile each expression within the body
		for _, expr := range n.Body {
			g.emitExpr(expr, child)
		}

	case *parser.Nil:
		g.emitln("    mov rax, 0       ; NIL")
		g.emitln("    TAG_NIL_REG rax  ; Tagged")

	case *parser.String:
		// create a label, based on the hash of the content.
		// This has the side-effect of interning.
		lbl := g.addString(n.Value)

		// save the string, because we're gonna put it into the
		// generated code, later.
		g.strings[lbl] = n.Value

		// load the address of the label and tag.
		g.emitln(fmt.Sprintf("    lea rax, %s", lbl))
		g.emitln("    TAG_STRING_REG rax")

	case *parser.Set:

		g.emitExpr(n.Expr, ev)

		if offset, ok := ev.Lookup(n.Name); ok {
			g.emitln(fmt.Sprintf(
				"    mov [rbp-%d], rax",
				offset,
			))
			return
		}

		if offset, ok := ev.LookupCapture(n.Name); ok {
			g.emitln(fmt.Sprintf(
				"    mov [r15+%d], rax",
				offset,
			))
			return
		}
		panic("unknown variable: " + n.Name)

	case *parser.Symbol:
		if offset, ok := ev.Lookup(n.Name); ok {
			g.emitln(fmt.Sprintf(
				"    mov rax, [rbp-%d]",
				offset,
			))
			return
		}

		if offset, ok := ev.LookupCapture(n.Name); ok {
			g.emitln(fmt.Sprintf(
				"    mov rax, [r15+%d]",
				offset,
			))
			return
		}
		panic("unknown symbol: " + n.Name)
	default:
		panic(fmt.Sprintf("emitExpr: Unhandled node type:%T value:%V\n", n, n))
	}
}

// emitDefun emits the body for the given function definition "(defun ..)".
//
// TODO: this is basically a copy/paste of emitLambda
func (g *Compiler) emitDefun(fn *parser.Defun) {

	g.emitln(g.asmName(fn.Name) + ":")

	g.emitln("    push rbp")
	g.emitln("    mov rbp, rsp")
	g.emitln("    sub rsp, 256 ;; guess at space for locals")

	ev := env.New(nil)

	regs := []string{
		"rdi",
		"rsi",
		"rdx",
		"rcx",
		"r8",
		"r9",
	}

	for i, p := range fn.Params {
		offset := (i + 1) * 8

		ev.Define(p, offset)

		g.emitln(fmt.Sprintf(
			"    mov [rbp-%d], %s",
			offset,
			regs[i],
		))
	}

	for _, xpr := range fn.Exprs {
		g.emitExpr(xpr, ev)
	}

	g.emitln("    leave")
	g.emitln("    ret")
}

// emitLambda emits the body for the given lambda definition "(lambda ..)".
//
// TODO: this is basically a copy/paste of emitDefun.
func (g *Compiler) emitLambda(l *parser.Lambda) {

	g.emitln(l.Name + ":")

	g.emitln("    push rbp")
	g.emitln("    mov rbp, rsp")
	g.emitln("    sub rsp, 256 ;; guess at space for locals")

	lambdaEnv := env.New(nil)

	regs := []string{
		"rdi",
		"rsi",
		"rdx",
		"rcx",
		"r8",
		"r9",
	}

	for i, p := range l.Params {
		offset := (i + 1) * 8

		lambdaEnv.Define(p, offset)

		g.emitln(fmt.Sprintf(
			"    mov [rbp-%d], %s",
			offset,
			regs[i],
		))
	}
	for i, cap := range l.Captures {
		lambdaEnv.DefineCapture(cap, 8*(i+1))
	}

	for _, xpr := range l.Exprs {
		g.emitExpr(xpr, lambdaEnv)
	}

	g.emitln("    leave")
	g.emitln("    ret")
}

// Compile creates and returns the assembly language source for the given
// list of functions.
func (g *Compiler) Compile(defs []*parser.Defun) (string, error) {

	// Ensure our string table is pristine
	g.strings = make(map[string]string)

	defuns := ""

	// Generate the user-defined functions to our internal buffer.
	for _, d := range defs {
		g.emitDefun(d)
		g.emitln("")
	}

	// Get them, and clear the buffer.
	defuns = g.text.String()
	g.text.Reset()

	// Now user-defined lambdas
	lambdas := ""
	for _, l := range g.lambdas {
		g.emitLambda(l)
		g.emitln("")
	}
	lambdas = g.text.String()
	g.text.Reset()

	// Then the string-table for user-defined strings
	stringTable := ""
	g.emitln("section .data")
	for id, str := range g.strings {
		g.emitln("align 8")
		g.emitln(id + ":")

		// escape the "`" which are wrapped around the string.
		str = strings.ReplaceAll(str, "`", "\\`")

		g.emitln(fmt.Sprintf("     db `%s`, 0", str))
	}
	stringTable = g.text.String()
	g.text.Reset()

	// Define a simple structure we can pass to the text/template
	// file we render for our output
	type Generated struct {
		// The defintions of defun's we've seen.
		Defuns string

		// Lambdas has all the lambda expressions we've seen.
		Lambdas string

		// StringTable contains the strings we've seen.
		StringTable string
	}

	// Create an instance to populate the template with
	x := &Generated{
		Defuns:      defuns,
		Lambdas:     lambdas,
		StringTable: stringTable,
	}

	// Create a buffer to render the template to.
	buf := bytes.Buffer{}

	// Load the template, and parse it.
	t1 := template.New("t1")
	t1 = template.Must(t1.Parse(tmplTxt))

	// Render the template.
	err := t1.Execute(&buf, x)
	if err != nil {
		return "", err
	}

	// Now return the text of that rendered template.
	return buf.String(), nil
}
