# Compile *.lisp in the current directory.
PROGRAMS := $(basename $(wildcard *.lisp))
.PHONY: test clean


# build the compiler
slisp: main.go
	go build .

# clean everything
clean:
	rm -f slisp $(PROGRAMS) *.asm *.o
	cd test     && make clean
	cd examples && make clean


#
# "make all" will compile all the *.lisp programs in the current
# directory; this is useful because I often have scratch programs
# present when experimenting.
#
all: $(PROGRAMS)

# generic rule to build a binary from a .lisp file
%: %.lisp slisp
	@echo "compiling $@"
	@./slisp -c $<

# Run functional tests
test:
	cd examples/    && make test
	cd examples/bf/ && make test
	cd test/        && make test
