#
# Makefile to run all the *.bf test programs through
# our interpreter and test their output matches that
# we expect
#


BF_FILES := $(wildcard *.bf)
OUT_FILES := $(BF_FILES:.bf=.bf.out)

all: $(OUT_FILES)
test: $(OUT_FILES)

#
# Ensure we have the interpreter
#
../brainfuck:
	$(MAKE) -C ..

#
# Generic rule, each program is run with a copy of
# its own source code on STDIN.
#
# This allows "cat.bf" to work, and is harmless for
# all the other programs (except factor.bf).
#
%.bf.out: %.bf ../brainfuck
	@echo "Running $<"
	@cat $< | ../brainfuck $< > $@
	@echo -n "Testing output .."
	@cmp $@ $<.expected || (echo "program $< did not match" ; cat $@.out ; exit 1)
	@echo "OK"

#
# Factor needs a number on STDIN
#
factor.bf.out: factor.bf ../brainfuck
	@echo "Running $<"
	@echo 120 | ../brainfuck $< > $@
	@echo -n "Testing output .."
	@cmp $@ $<.expected || (echo "program $< did not match" ; cat $@.out ; exit 1)
	@echo "OK"


clean:
	rm -f $(OUT_FILES)
