PROGRAMS := $(basename $(wildcard *.lisp))
TESTS := $(addsuffix .test,$(PROGRAMS))

.PHONY: all clean

all: $(PROGRAMS)

clean:
	@rm -f *.s *.o *.out *.test tmp.asm $(PROGRAMS)

test: $(TESTS)

# Ensure we have the compiler
../slisp:
	cd .. && go build .


# Generic rule
#   "foo.lisp" becomes "foo"
#   "foo" is executed to produce "foo.out"
#   "foo.out" is compared to "foo.expected" which is the expected output
#
%: %.lisp ../slisp
	@echo "compiling $@"
	@../slisp $< > tmp.asm
	@nasm -f elf64 tmp.asm
	@ld -o $@ tmp.o

# Generic rule for testing each binary
%.test: %.lisp ../slisp
	@echo "compiling $@"
	@../slisp $< > tmp.asm
	@nasm -f elf64 tmp.asm
	@ld -o $@ tmp.o
	@./$@ > $@.out 2>&1  || true
	@cmp $@.out $@.expected || (echo "program $< did not match" ; cat $@.out ; exit 1)
