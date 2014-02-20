package tpl

import (
	"bytes"
	"io"
	"io/ioutil"
	path "path/filepath"
	"regexp"
	"strings"
	"text/template"
)

// Some helpful regular expressions to control whitespace.
var (
	stripNewLines     = regexp.MustCompile("}}\n")
	stripLeadingSpace = regexp.MustCompile("(?m)^(\t| )+") // multi-line mode
	stripTooManyLines = regexp.MustCompile("\n\n\n+")
)

// ParseText will open the given file path and parse its contents as a
// template. If fpath has length 0, then a default set of templates is used.
//
// Note that ParseText does some pre-processing on the template text as a way
// to reasonably control whitespace. Namely, lines ending with '}}' are
// completely ignored. Lines ending with '}}\' are not ignored. The template
// is parsed before these changes are made so that accurate line numbers can
// be given in case there is an error.
func ParseText(fpath string) (*template.Template, error) {
	var tname, text string
	if len(fpath) == 0 {
		tname = "default"
		text = defaults
	} else {
		tname = path.Base(fpath)
		bs, err := ioutil.ReadFile(fpath)
		if err != nil {
			return nil, ef("Could not read '%s': %s", fpath, err)
		}
		text = string(bs)
	}

	// Try to parse the templates before mangling them, so that error
	// messages retain meaningful line numbers.
	_, err := template.New(tname).Funcs(Functions).Parse(text)
	if err != nil {
		return nil, ef("Problem parsing template: %s", err)
	}

	// Okay, now do it for real.
	text = trimTemplate(text)
	t, err := template.New(tname).Funcs(Functions).Parse(text)
	if err != nil {
		return nil, ef("BUG: Problem parsing template: %s", err)
	}
	return t, nil
}

// ExecText performs a standard template exec, except it does some
// post-processing on the output to control whitespace. Namely, 3 or more
// consecutive new line characters (LF) are replaced with 2 new line
// characters (LF).
func ExecText(t *template.Template, w io.Writer, data interface{}) error {
	buf := new(bytes.Buffer)
	if err := t.Execute(buf, data); err != nil {
		return err
	}
	out := stripTooManyLines.ReplaceAllString(buf.String(), "\n\n")
	_, err := io.WriteString(w, out)
	return err
}

func trimTemplate(s string) string {
	// Order is important here.
	s = stripLeadingSpace.ReplaceAllString(s, "")
	s = stripNewLines.ReplaceAllString(s, "}}")
	s = strings.Replace(s, "}}\\", "}}", -1)
	return s
}
