export FOO=BAR
export NAME=Steve

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



# Optionally remove unused sections and strip the binaries.
ifeq ($(SMALL),1)
LINK_CMD = @ld --gc-sections -s -o $@ tmp.o
else
LINK_CMD = @ld -o $@ tmp.o
endif


# Generic rule
#   "foo.lisp" becomes "foo"
#   "foo" is executed to produce "foo.out"
#   "foo.out" is compared to "foo.expected" which is the expected output
#
%: %.lisp ../slisp
	@echo "compiling $@"
	@../slisp $< > tmp.asm
	@nasm -f elf64 tmp.asm
	$(LINK_CMD)

# Life isn't gonna get tested
life.test:
	@echo "ignoring $@ - as it doesn't terminate"

wc.test:
	@echo "ignoring $@ - as it doesn't terminate"

# Generic rule for testing each binary
%.test: %.lisp ../slisp
	@echo "compiling $@"
	@../slisp $< > tmp.asm
	@nasm -f elf64 tmp.asm
	$(LINK_CMD)
	@echo "testing $@"
	@./$@ > $@.out 2>&1  || true
	@cmp $@.out $@.expected || (echo "program $< did not match" ; cat $@.out ; exit 1)
