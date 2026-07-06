// trivial lisp compiler which generates nasm-style assembly.

package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/skx/slisp/compiler"
)

//go:embed stdlib.slisp
var stdlibLisp string

// generate is a helper to compile a program
func generate(prg string) (string, error) {

	// Create a compiler
	c := compiler.New(prg)

	// Generate the code
	txt, err := c.Compile()
	if err != nil {
		return "", fmt.Errorf("error compiling program %s", err)
	}
	return txt, nil
}

// compile will compile the given program into an object,
// then a binary.
func compile(name string, txt string) {

	// Get the basename
	nm := filepath.Base(name)

	// Remove the .lisp suffix
	nm = strings.TrimSuffix(nm, filepath.Ext(nm))

	// write the assembly
	err := os.WriteFile(nm+".asm", []byte(txt), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to write assembly to %s: %s\n", nm+".asm", err)
		os.Exit(1)
	}

	// nasm
	assembleCmd := []string{"nasm", "-f", "elf64", nm + ".asm"}

	c := exec.Command(assembleCmd[0], assembleCmd[1:]...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	err = c.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error running assembler %v: %s\n", assembleCmd, err)
		os.Exit(1)
	}

	// link
	linkCmd := []string{"ld", "-o", nm, "--gc-sections", "-s", nm + ".o"}

	c = exec.Command(linkCmd[0], linkCmd[1:]...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	err = c.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error running linker %v: %s\n", linkCmd, err)
		os.Exit(1)
	}
}

// main is our entry-point
func main() {

	// CLI flags
	stdlibFlag := flag.Bool("stdlib", true, "Prepend our Lisp standard library to user-programs.")
	compileFlag := flag.Bool("compile", false, "Automatically generate a binary.")
	c := flag.Bool("c", false, "Automatically generate a binary.")
	flag.Parse()

	// Do we have a file?
	if len(flag.Args()) != 1 {
		fmt.Println("usage: slisp [flags] file.lisp")
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
	if *stdlibFlag {
		prg = stdlibLisp + "\n" + prg
	}

	txt, err := generate(prg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error processing: %s\n", err)
		os.Exit(1)
	}

	if *c || *compileFlag {
		compile(flag.Args()[0], txt)
	} else {
		// Print the code to STDOUT
		fmt.Print(txt)
	}
}
