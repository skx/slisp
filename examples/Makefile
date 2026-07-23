export FOO=BAR
export TEST=1
export NAME=Steve

PROGRAMS := $(basename $(wildcard *.lisp))
TESTS := $(addsuffix .test,$(PROGRAMS))

.PHONY: all clean

all: $(PROGRAMS)

clean:
	@rm -f *.s *.o *.out *.test tmp.asm $(PROGRAMS)
	@cd bf/ && make clean

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
	@../slisp -c $<

# inception needs a file to be named
inception.test: ../slisp
	@echo "compiling $@"
	@../slisp -c -o inception.test inception.lisp
	@echo "testing $@"
	@./$@ inception.in > $@.out 2>&1  || true
	@cmp $@.out $@.expected || (echo "program $< did not match" ; cat $@.out ; exit 1)

wc.test:
	@echo "ignoring $@ - as it doesn't terminate"

life.test:
	@echo "ignoring $@ - as it doesn't terminate"

# Generic rule for testing each binary
%.test: %.lisp ../slisp
	@echo "compiling $@"
	@../slisp -c -o $@ $<
	@echo "testing $@"
	@./$@ > $@.out 2>&1  || true
	@cmp $@.out $@.expected || (echo "program $< did not match" ; cat $@.out ; exit 1)
