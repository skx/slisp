(defun read (file)
  "Read and return the contents of the named file,
returning NIL on failure.

   The fread, fwrite, and fclose functions each allow
a NIL file-handle to be specified to allow this simple
pattern of using a (let ..) scope to handle the I/O."
  (let ((handle (fopen file "r"))  ; open
        (data   (fread handle))    ; read
        (res    (fclose handle)))  ; close
    data))

(defun main ()
  (print "Reading a missing file: ")
  (println (read "/path/does/not/exist!"))
  (newline)

  (println "Reading our own source code.")
  (print (read "fread.lisp"))
  (newline))
