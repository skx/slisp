slisp: main.go
	go build .

clean:
	rm -f slisp example example.o example.s

test: slisp example.lisp
	./slisp example.lisp  > example.s
	nasm -f elf64 example.s
	ld -o example example.o
	./example
