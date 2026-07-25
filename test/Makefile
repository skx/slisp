export FOO=BAR
export TEST=TEST
export NAME=Steve

PROGRAMS := $(basename $(wildcard *.lisp))
TESTS := $(addsuffix .test,$(PROGRAMS))

.PHONY: all clean test test-inception

all: $(PROGRAMS)

clean:
	@rm -f *.asm *.o *.out *.test *.itest tmp.asm $(PROGRAMS)

#
# Compile each test, and compare the results of executing it with expected ones.
#
test: ../slisp
	@$(MAKE) RUN='./$$@' $(TESTS)

#
# Compile each test, but ignore the result.
#
# Instead execute the test-case via the interpreter and compare the results with expected ones.
#
test-inception: ../examples/inception
	@$(MAKE) RUN='../examples/inception $$< --eval="(main (list \"a.out\"))"' $(TESTS)



# Ensure we have the compiler
../slisp:
	cd .. && go build .


# Ensure we have inception
../examples/inception:
	cd ../examples && make inception


# Generic rule to compile "foo.lisp" into "foo" and run tests
#
# NOTE: The RUN command is set so it might be that we don't
#       even need the compiled program.
%.test: %.lisp ../slisp
	@echo "Running test for $@"
	@../slisp -c -o $@ $<
	@$(RUN) > $@.out 2>&1 || true
	@cmp $@.out $@.expected || \
		(echo "program $< did not match"; cat $@.out; exit 1)
