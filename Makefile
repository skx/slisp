.PHONY: test clean

slisp: main.go
	go build .

clean:
	rm -f slisp example example.o example.s
	cd test && make clean

example: slisp example.lisp
	./slisp example.lisp  > example.s
	nasm -f elf64 example.s
	ld -o example example.o
	./example

test:
	cd test && make test
