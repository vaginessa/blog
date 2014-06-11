package main

import (
	"bytes"
	"fmt"
	"regexp"

	"github.com/kjk/u"
)

var LANG_TO_PRETTIFY_LANG_MAP = map[string]string{
	"c":          "c",
	"c++":        "cc",
	"cpp":        "cpp",
	"python":     "py",
	"html":       "html",
	"xml":        "xml",
	"perl":       "pl",
	"c#":         "cs",
	"javascript": "js",
	"java":       "java",
}

func lang_to_prettify_lang(lang string) string {
	//from http://google-code-prettify.googlecode.com/svn/trunk/README.html
	//"bsh", "c", "cc", "cpp", "cs", "csh", "cyc", "cv", "htm", "html",
	//"java", "js", "m", "mxml", "perl", "pl", "pm", "py", "rb", "sh",
	//"xhtml", "xml", "xsl"
	if l, ok := LANG_TO_PRETTIFY_LANG_MAP[lang]; ok {
		return fmt.Sprintf("lang-%s", l)
	}
	return ""
}

func txt_cookie(s string) string {
	return u.Sha1StringOfBytes([]byte(s))
}

var reCode = regexp.MustCompile("(?siU)<code.*>.+</code>")

func extractLang(s []byte) (rest, lang []byte) {
	for len(s) > 0 {
		if s[0] != ' ' {
			break
		}
		s = s[1:]
	}
	i := bytes.IndexByte(s, '>')
	if 0 == i {
		return s[1:], nil
	}
	return s[i+1:], s[:i]
}

func txt_with_code_parts(txt []byte) ([]byte, map[string][]byte) {
	code_parts := make(map[string][]byte)
	res := reCode.ReplaceAllFunc(txt, func(s []byte) []byte {
		s = s[len("<code") : len(s)-len("</code>")]
		s, lang := extractLang(s)
		new_code := ""
		if lang != nil {
			l := lang_to_prettify_lang(string(lang))
			new_code = fmt.Sprintf(`<pre class="prettyprint %s">%s</pre>`, l, string(s))
		} else {
			new_code = fmt.Sprintf(`<pre class="prettyprint">%s</pre>`, string(s))
		}
		cookie := txt_cookie(new_code)
		code_parts[cookie] = []byte(new_code)
		return []byte(cookie)
	})
	return res, code_parts
}
