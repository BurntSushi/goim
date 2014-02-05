package tpl

import (
	"strings"
	"text/template"

	"github.com/kr/text"
)

var Helpers = template.FuncMap{
	"combine": HelpCombine,
	"lines":   HelpLines,
	"wrap":    HelpWrap,
}

// HelpCombine provides a way to compose values during template execution.
// This is particularly useful when executing sub-templates. For example,
// say you've defined two variables `$a` and `$b` that you want to pass to
// a sub-template. But templates can only take a single pipeline. Combine will
// let you bind any number of values. For example:
//
//	{{ template "tpl_name" (Combine "a" $a "b" $b) }}
//
// The template "tpl_name" can then access `$a` and `$b` with `.a` and `.b`.
//
// Note that the first and every other subsequent value must be strings. The
// second and every other subsequent value may be anything. There must be an
// even number of arguments given. If any part of this contract is violated,
// the function panics.
func HelpCombine(keyvals ...interface{}) map[string]interface{} {
	if len(keyvals)%2 != 0 {
		panic(sf("Combine must have even number of parameters but %d isn't.",
			len(keyvals)))
	}
	m := make(map[string]interface{})
	for i := 0; i < len(keyvals); i += 2 {
		key, ok := keyvals[i].(string)
		if !ok {
			panic(sf("Parameter %d to Combine must be a string but it is "+
				"a %T.", i, keyvals[i]))
		}
		m[key] = keyvals[i+1]
		// pef("### %#v :: %#v", keyvals[i], keyvals[i+1])
	}
	return m
}

func HelpWrap(limit int, s interface{}) string {
	return text.Wrap(sf("%s", s), limit)
}

func HelpLines(s interface{}) []string {
	return strings.Split(sf("%s", s), "\n")
}
