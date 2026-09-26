;; Test making syscalls

(defun main(args)

  ;; write to stdout
  (syscall 1 1 "STDOUT\n" 7)

  ;; write to stderr
  (syscall 1 2 "ERROR\n"  6)

  ;; exit with return code 31.
  (syscall 60 31)

)
