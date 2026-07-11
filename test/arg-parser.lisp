(require arg-parser)

(defun dump (args)
  "Create a parser with the given args, and dump the output it produces."
  (let ((parser (arg-parser:new args)))
    (println "Created parser with arguments: " args)
    (println "\tArgs :" (parser :files))
    (println "\tFlags:" (parser :flags))))


(defun main()
  (let ((a1 (list "--long" "/etc/passwd"))
        (a2 (list "-lwc" "id")))

    ;; obvious cases
    (dump a1)
    (dump a2)

    ;; inline
    (dump (list "foo" "bar" "baz" "-w" "-l" "--long=-s"))

    ;; test empty flag
    (dump (list "-"))

    ;; test empty list.
    (dump (list))
    ))
