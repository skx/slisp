(defun read (file)
  "Read and return the contents of the named file,
 return NIL on failure to open, or read."
  (let ((handle (fopen file "r")))
    (if handle
        (let ((data (fread handle)))
          (fclose handle)
          data))))

(defun write(file str)
  "Write the string to the given file"
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
