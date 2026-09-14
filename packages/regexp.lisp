;;; simple regular expression functions.
;;
;; We support only basic facilities here:  ".", "?", and "*",
;; along with the two anchors "^" and "$".
;;

(defun string-match-char? (pattern text)
  "Does the pattern character match the given text character?"
  (or (= pattern #\.)
      (= pattern text)))


(defun string-match-helper (pattern text)
  "Try to match PATTERN at the beginning of TEXT."

  ;; We've matched the entire pattern.  This is success even if
  ;; there is remaining text.  '$' is handled explicitly below.
  (if (nil? pattern)
      t

      (let ((p (car pattern))
            (rest (cdr pattern)))

        ;; '$' means the match must end here.
        (if (= p #\$)
            (if (nil? rest)
                (nil? text)
                nil)

            ;; '^' is handled by string-match, not here.
            (if (= p #\^)
                nil

                ;; '*' applies to the preceding character.
                (if (and rest (= (car rest) #\*))

                    ;; Either:
                    ;;   1. '*' matches zero characters, or
                    ;;   2. '*' consumes one matching character.
                    (if (nil? text)
                        (string-match-helper
                         (cdr rest)
                         text)

                        (if (string-match-char? p (car text))
                            (or
                             (string-match-helper
                              pattern
                              (cdr text))
                             (string-match-helper
                              (cdr rest)
                              text))
                            (string-match-helper
                             (cdr rest)
                             text)))

                    ;; '?' applies to the preceding character.
                    (if (and rest (= (car rest) #\?))

                        (or
                         ;; Consume the optional character.
                         (if text
                             (if (string-match-char? p (car text))
                                 (string-match-helper
                                  (cdr rest)
                                  (cdr text))
                                 nil)
                             nil)

                         ;; Or don't consume it.
                         (string-match-helper
                          (cdr rest)
                          text))

                        ;; Ordinary character / '.'
                        (if text
                            (if (string-match-char? p (car text))
                                (string-match-helper
                                 rest
                                 (cdr text))
                                nil)
                            nil))))))))


(defun string-match-search (pattern text)
  "Try PATTERN at every possible starting position in TEXT."

  (if (string-match-helper pattern text)
      t
      (if (nil? text)
          nil
          (string-match-search
           pattern
           (cdr text)))))


(defun regexp:match (pattern text)
  "Return true if PATTERN matches anywhere in TEXT."

  (let ((p (explode pattern))
        (s (explode text)))

    ;; '^' anchors the match to the beginning.
    (if (and p (= (car p) #\^))
        (string-match-helper (cdr p) s)

        ;; Otherwise, try every possible starting position.
        (string-match-search p s))))
