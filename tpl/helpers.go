package tpl

import (
	"reflect"
	"strings"
	"text/template"

	"github.com/kr/text"

	"github.com/BurntSushi/goim/imdb"
)

var Helpers = template.FuncMap{
	"combine":    combine,
	"lines":      lines,
	"wrap":       wrap,
	"underlined": underlined,

	"plots":  attrGetter(new(imdb.Plots)),
	"quotes": attrGetter(new(imdb.Quotes)),
	"rank":   attrGetter(new(imdb.UserRank)),
}

// Combine provides a way to compose values during template execution.
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
func combine(keyvals ...interface{}) map[string]interface{} {
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
	}
	return m
}

func wrap(limit int, s interface{}) string {
	return text.Wrap(sf("%s", s), limit)
}

func lines(s interface{}) []string {
	return strings.Split(sf("%s", s), "\n")
}

func underlined(rep string, is interface{}) string {
	s := sf("%s", is)
	return sf("%s\n%s", s, strings.Repeat(rep, len(s)))
}

// attrGetter does some fancy reflection footwork to automatically build
// a function for any attribute retriever satisfying the imdb.Attributer
// interface.
func attrGetter(attrs imdb.Attributer) interface{} {
	// So we can make new attrs values.
	// Note that this is the *underlying* type of the imdb.Attributer.
	tattrs := reflect.TypeOf(attrs).Elem()
	return func(e imdb.Entity) interface{} {
		assertDB()
		vattrs := reflect.New(tattrs).Interface().(imdb.Attributer)
		assert(e.Attrs(tplDB, vattrs))
		return vattrs
	}
}
