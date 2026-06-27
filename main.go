// trivial lisp compiler which generates nasm-style assembly.

package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"

	"github.com/skx/slisp/compiler"
)

//go:embed stdlib.slisp
var stdlibLisp string

// compile is a helper to compile a program
func compile(prg string) (string, error) {

	// Create a compiler
	c := compiler.New(prg)

	// Generate the code
	txt, err := c.Compile()
	if err != nil {
		return "", fmt.Errorf("error compiling program %s", err)
	}
	return txt, nil
}

// main is our entry-point
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

	// Prepend the stdlib if we should.
	prg := string(data)
	if *stdlib {
		prg = stdlibLisp + "\n" + prg
	}

	txt, err := compile(prg)
	if err != nil {
		fmt.Printf("error processing: %s\n", err)
		return
	}

	// Print the code to STDOUT
	fmt.Print(txt)
}
