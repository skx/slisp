# Compile *.lisp in the current directory.
PROGRAMS := $(basename $(wildcard *.lisp))
.PHONY: test clean


# build the compiler
slisp: main.go
	go build .


# Example should run after being built
example: example.lisp slisp
	@echo "compiling $@"
	@./slisp $< > $@.asm
	@nasm -f elf64 $@.asm
	@ld -o $@ $@.o
	@./example

# clean everything
clean:
	rm -f slisp $(PROGRAMS) *.asm *.o
	cd test && make clean


#
# "make all" will compile all the *.lisp programs in the current
# directory; this is useful because I often have scratch programs
# present when experimenting.
#
all: $(PROGRAMS)

# generic rule to build a binary from a .lisp file
%: %.lisp slisp
	@echo "compiling $@"
	@./slisp $< > $@.asm
	@nasm -f elf64 $@.asm
	@ld -o $@ $@.o

# Run functional tests
test:
	cd test && make test
