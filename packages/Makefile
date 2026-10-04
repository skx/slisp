nop:
	@echo "Run make syscall to update syscall.lisp"

syscall:
	{ cat syscall.header; awk '/^#define __NR_/ { name=$$2; sub(/^__NR_/, "", name); print "(defconst SYS_" name " " $$3 ")" }' /usr/include/x86_64-linux-gnu/asm/unistd_64.h; } > syscall.lisp
