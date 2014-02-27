/*
Package tpl provides convenience functions that are loaded into every Goim
template, along with some functions for parsing and executing Goim templates.
This package also defines a few key types, like Args, that describe all values
passed to a template when executed.

In general, every template executed targets one specific entity, which is
set in the E field of the Args struct. Some templates require additional
information, which is set in the A field of the Args struct.

This package also uses global state to set the database to use in some of the
helper functions defined. This is unfortunate but convenient. If SetDB is not
called and a helper function is used that requires it, executing the template
will return an error.

Beta

Please consider this package as beta material. I am reasonably happy with what
is here so far and I don't expect to change it. But it hasn't been used much,
so I'd like to wait before declaring it stable.
*/
package tpl
