;;; simple regular expression functions.
;;
;; We support only basic facilities here:
;;
;;  .  Any single character
;;  ?  Zero or one occurrence of the previous character/set
;;  *  Zero or more occurrences of the previous character/set
;;  +  One or more occurrences of the previous character/set
;;
;; Sets allow any character in the class to be matched, so
;; [abc] matches either a, b, or c, and "[abc]*" could match
;; "abc", "aaaaaac", "bac", or similar.
;;
;; along with character-classes and the anchoring characters
;; "^" and "$".
;;

(defun string-match-char? (pattern text)
  "Does the pattern character match the given text character?"

  ;; A character class is represented by a list of characters.
  (if (cons? pattern)

      ;; Does TEXT occur in the character class?
      (if (nil? pattern)
          nil
          (or (= (car pattern) text)
              (string-match-char? (cdr pattern) text)))

      ;; Ordinary character or '.'
      (or (= pattern #\.)
          (= pattern text))))


(defun regexp-tokenize (pattern)
  "Convert the exploded pattern into pattern tokens.

A character class such as [abc] becomes one token containing
the list (#\a #\b #\c)."

  (if (nil? pattern)
      nil

      (let ((p (car pattern))
            (rest (cdr pattern)))

        (if (= p #\[)

            ;; Read characters until ']'.
            (let ((class (regexp-read-class rest)))
              (cons (car class)
                    (regexp-tokenize (cdr class))))

            (cons p
                  (regexp-tokenize rest))))))


(defun regexp-read-class (pattern)
  "Read a character class.

Return (CLASS REMAINING-PATTERN)."

  (if (nil? pattern)
      ;; Unterminated character class.
      ;; Treat '[' as an ordinary character.
      (cons (list #\[) nil)

      (let ((p (car pattern))
            (rest (cdr pattern)))

        (if (= p #\])
            (cons nil rest)

            (let ((result (regexp-read-class rest)))
              (cons (cons p (car result))
                    (cdr result)))))))


(defun string-match-helper (pattern text)
  "Try to match PATTERN at the beginning of TEXT."

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

                ;; '*' applies to the preceding character/set.
                (if (and rest (= (car rest) #\*))

                    (if (nil? text)
                        (string-match-helper
                         (cdr rest)
                         text)

                        (if (string-match-char? p (car text))
                            (or
                             ;; Consume another character.
                             (string-match-helper
                              pattern
                              (cdr text))
                             ;; Stop consuming.
                             (string-match-helper
                              (cdr rest)
                              text))
                            ;; '*' can match zero characters.
                            (string-match-helper
                             (cdr rest)
                             text)))

                    ;; '+' applies to the preceding character/set
                    ;; and requires at least one match.
                    (if (and rest (= (car rest) #\+))

                        (if (nil? text)
                            nil

                            (if (string-match-char? p (car text))
                                (or
                                 ;; Consume another character.
                                 (string-match-helper
                                  pattern
                                  (cdr text))
                                 ;; We've consumed at least one, so
                                 ;; continue with the rest of pattern.
                                 (string-match-helper
                                  (cdr rest)
                                  (cdr text)))
                                nil))

                        ;; '?' applies to the preceding character/set.
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

                            ;; Ordinary character / '.' / set.
                            (if text
                                (if (string-match-char? p (car text))
                                    (string-match-helper
                                     rest
                                     (cdr text))
                                    nil)
                                nil)))))))))

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

  (let ((p (regexp-tokenize (explode pattern)))
        (s (explode text)))

    ;; '^' anchors the match to the beginning.
    (if (and p (= (car p) #\^))
        (string-match-helper (cdr p) s)

        ;; Otherwise, try every possible starting position.
        (string-match-search p s))))
