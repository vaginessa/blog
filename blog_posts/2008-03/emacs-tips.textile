Id: 1001
Title: Emacs tips.
Tags: emacs
Date: 2008-03-13T07:24:18-07:00
Format: Markdown
--------------
In `.bashrc`:

<code>\
alias emacs=emacs -nw —no-splash\
</code>

In `.emacs`:

<code>\
(setq-default transient-mark-mode t)\
(global-set-key "\\M-g&quot; 'goto-line)\
(global-set-key "\\C-z&quot; 'undo)

;; Put autosave files (ie \#foo\#) in one place, **not** scattered all
over the\
;; file system! (The make-autosave-file-name function is invoked to
determine\
;; the filename of an autosave file.)\
(defvar autosave-dir "\~/.emacs.bak/&quot;)\
(make-directory autosave-dir t)\
(defun auto-save-file-name-p (filename) (string-match "\^\#.\*\#\$&quot;
(file-name-nondirectory filename)))

(defun make-auto-save-file-name ()\
 (concat autosave-dir\
 (if buffer-file-name\
 (concat "\#" (file-name-nondirectory buffer-file-name) "\#")\
 (expand-file-name\
 (concat "\#%" (buffer-name) "\#")\
 ))))

;; Put backup files (ie foo\~) in one place too. (The
backup-directory-alist\
;; list contains regexp=\>directory mappings; filenames matching a
regexp are\
;; backed up in the corresponding directory. Emacs will mkdir it if
necessary.)\
(defvar backup-dir "\~/.emacs.bak/&quot;)\
(setq backup-directory-alist (list (cons "." backup-dir)))

</code>
