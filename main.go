// trivial lisp compiler which generates nasm-style assembly.

package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"

	"github.com/skx/slisp/compiler"
	"github.com/skx/slisp/parser"
)

//go:embed stdlib.slisp
var stdlibLisp string

// main
func main() {

	// CLI flags
	stdlib := flag.Bool("stdlib", true, "Prepend our Lisp standard library to user-programs")
	flag.Parse()

	// Do we have a file?
	if len(flag.Args()) != 1 {
		fmt.Println("usage: slisp [-stdlib=false] file.lisp")
		os.Exit(1)
	}

	// Read the file-contents
	data, err := os.ReadFile(flag.Args()[0])
	if err != nil {
		fmt.Printf("failed to read input %s: %s\n", os.Args[1], err)
		return
	}

	// Append the stdlib if we should.
	prg := string(data)
	if *stdlib {
		prg = stdlibLisp + "\n" + prg
	}

	// Create a parser
	p := parser.New(prg)

	// Parse into functions
	defs, err := p.Parse()
	if err != nil {
		fmt.Printf("error parsing program %s\n", err)
		return
	}

	// Create a compiler
	c := compiler.New()

	// Generate the code
	txt := ""
	txt, err = c.Compile(defs)
	if err != nil {
		fmt.Printf("Error compiling program %s\n", err)
		return
	}

	// Print the code to STDOUT
	fmt.Print(txt)
}
