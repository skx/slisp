(defun read (file)
  "Read and return the contents of the named file,
returning NIL on failure.

   The fread, fwrite, and fclose functions each allow
a NIL file-handle to be specified to allow this simple
pattern of using a (let ..) scope to handle the I/O."
  (let ((handle  (fopen file "r"))    ; open
        (result  (fread handle))      ; read
        (discard (fclose handle)))    ; close
    result))

(defun write(file str)
  "Write the string to the given file.

Similar story here with regard to the fwrite/close on
a file-handle that might be nil."
  (let ((handle  (fopen file "w"))                 ; open
        (result  (fwrite handle str (length str))) ; write
        (discard (fclose handle)))                 ; close
    result))

(defun main ()
  ; write
  (print "Writing to a file: ")
  (println (write "test.txt" "Hello, world!"))

  ; read-back to confirm it worked.
  (println "Reading our freshly created/written file.")
  (print (read "test.txt"))
  (newline)

  ; cleanup
  (unlink "test.txt"))
