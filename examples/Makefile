export FOO=BAR
export TEST=1
export NAME=Steve

PROGRAMS := $(basename $(wildcard *.lisp))
TESTS := $(addsuffix .test,$(PROGRAMS))

.PHONY: all clean test test-inception


all: $(PROGRAMS)


clean:
	@rm -f *.asm *.o *.out *.test $(PROGRAMS)
	@cd bf/ && make clean


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
test-inception: ./inception
	@$(MAKE) RUN='./inception $$< --eval="(main (list \"./example.test\"))"' $(TESTS)


# Ensure we have the compiler
../slisp:
	cd .. && go build .


# Ensure we have inception
inception: inception.lisp ../slisp
	@../slisp -c $<




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
	@$(RUN) > $@.out 2>&1 || true
	@cmp $@.out $@.expected || (echo "program $< did not match" ; cat $@.out ; exit 1)
