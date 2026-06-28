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
  (let ((handle (fopen file "w")))
    (if handle
        (let ((result (fwrite handle str (length str))))
          (fclose handle)
          result))))

(defun main ()
  (print "Writing to a file: ")
  (println (write "test.txt" "Hello, world!"))

  (println "Reading our freshly created/written file.")
  (print (read "test.txt"))
  (newline))
