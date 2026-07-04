(defun delete(name)
  (if (exists? name)
      (do
       (println "file exists:" name)
       (unlink name))
      (println "file does not exist:" name)))

(defun main()
  ;; delete if present
  (delete "unlink.txt")

  ;; create the file
  (let ((handle (fopen "unlink.txt" "w"))
        (res    (fwrite handle "data" 4))
        (discard (fclose handle))))

  ;; now delete it
  (delete "unlink.txt"))
