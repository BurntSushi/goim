package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
	"text/template"

	"github.com/BurntSushi/ty/fun"

	"github.com/BurntSushi/xdg"
)

var flagConfigOverwrite = false

type config struct {
	Driver     string
	DataSource string `toml:"data_source"`
}

var defaultConfig = `
# The 'driver' is the type of relational database that you're using.
# Currently, goim has only been tested/optimized for SQLite and PostgreSQL.
# For SQLite, the driver name is 'sqlite3'.
# For PostgreSQL, the driver name is 'postgres'.
driver = "sqlite3"

# The data source specifies which database to connect to. For SQLite, this
# is simply a file path. If it's a relative file path, then it's interpreted
# with respect to the current working directory of wherever 'goim' is executed.
#
# If you're using a different relational database system, like PostgreSQL,
# then you will need to consult its documentation for specifying connection
# strings. For PostgreSQL, see: http://goo.gl/kKaxAj
#
# Here's an example PostgreSQL connection string:
#
#     user=andrew password=XXXXXX dbname=imdb sslmode=disable
#
# N.B. The 'sslmode=disable' appears to be required for a default PostgreSQL
# installation. (At least on Archlinux, anyway.)
data_source = "goim.sqlite"
`

// The default templates to write to the configuration directory.
// Note that each template has '{{define "..."}}...{{end}}' automatically
// added based on its name in the map.
// All leading and trailing whitespace is stripped from the templates provided
// here.
var defaultTpls = map[string]string{
	"info_movie": `
{{.Title}} ({{.Year}})

ID: {{.Id}}
`,
}

var xdgPaths = xdg.Paths{XDGSuffix: "goim"}

var cmdWriteConfig = &command{
	name:            "write-config",
	positionalUsage: "[ dir ]",
	shortHelp:       "write a default configuration",
	help: `
Writes the default configuration to $XDG_CONFIG_HOME/goim or to
the directory argument given.

If no argument is given and $XDG_CONFIG_HOME is not set, then the configuration
is written to $HOME/.config/goim/.

The configuration includes a TOML file for specifying database connection
parameters, along with a set of template files used to control the various
output formats of Goim.
`,
	flags: flag.NewFlagSet("write-config", flag.ExitOnError),
	run:   writeConfig,
	addFlags: func(c *command) {
		c.flags.BoolVar(&flagConfigOverwrite, "overwrite", flagConfigOverwrite,
			"When set, the config file will be written regardless of\n"+
				"whether one exists or not.")
	},
}

func writeConfig(c *command) {
	var dir string
	if arg := strings.TrimSpace(c.flags.Arg(0)); len(arg) > 0 {
		dir = arg
	} else {
		dir = strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME"))
		if len(dir) == 0 {
			dir = path.Join(os.Getenv("HOME"), ".config")
		}
		dir = path.Join(dir, "goim")
		if err := os.MkdirAll(dir, 0777); err != nil {
			fatalf("Could not create '%s': %s", dir, err)
		}
	}

	confPath := path.Join(dir, "config.toml")
	tplPath := path.Join(dir, "format.tpl")

	// Don't clobber the user's config unexpectedly!
	if !flagConfigOverwrite {
		_, err := os.Stat(confPath)
		if !os.IsNotExist(err) {
			fatalf("Config file at '%s' already exists. Remove or use "+
				"-overwrite.", confPath)
		}
		_, err = os.Stat(tplPath)
		if !os.IsNotExist(err) {
			fatalf("Template file at '%s' already exists. Remove or use "+
				"-overwrite.", tplPath)
		}
	}

	conf := []byte(strings.TrimSpace(defaultConfig) + "\n")
	if err := ioutil.WriteFile(confPath, conf, 0666); err != nil {
		fatalf("Could not write '%s': %s", confPath, err)
	}

	tpl, err := os.Create(tplPath)
	if err != nil {
		fatalf("Could not create '%s': %s", tplPath, err)
	}

	// Sort the names so we can deterministic output.
	tplNames := fun.Keys(defaultTpls).([]string)
	sort.Strings(tplNames)
	define, prefix := `%s{{ define "%s" }}%s{{ end }}`, ""
	for _, name := range tplNames {
		t := strings.TrimSpace(defaultTpls[name])
		_, err := fmt.Fprintf(tpl, define, prefix, name, t)
		if err != nil {
			fatalf("Could not write '%s': %s", tplPath, err)
		}
	}
}

func defaultTemplate(name string) *template.Template {
	tpl, ok := defaultTpls[name]
	if !ok {
		fatalf("BUG: No template with name '%s' exists.", name)
	}
	tpl = strings.TrimSpace(tpl)
	text := sf(`{{ define "%s" }}%s{{ end }}`, name, tpl)
	t, err := template.New(name).Parse(text)
	if err != nil {
		fatalf("BUG: Could not parse template '%s': %s", name, err)
	}
	return t
}
