package main

import (
	"flag"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/BurntSushi/xdg"

	"github.com/BurntSushi/goim/tpl"
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

var xdgPaths = xdg.Paths{XDGSuffix: "goim"}

var cmdWrite = &command{
	name:            "write",
	positionalUsage: "(config | templates) [ dir ]",
	shortHelp:       "write default configuration or templates",
	help: `
Writes the default configuration/templates to $XDG_CONFIG_HOME/goim or to
the directory argument given.

If no argument is given and $XDG_CONFIG_HOME is not set, then the configuration
is written to $HOME/.config/goim/.

The configuration is a TOML file for specifying database connection
parameters, and the templates control the output formats of Goim on the command
line.
`,
	flags: flag.NewFlagSet("write", flag.ExitOnError),
	run:   cmd_write,
	addFlags: func(c *command) {
		c.flags.BoolVar(&flagConfigOverwrite, "overwrite", flagConfigOverwrite,
			"When set, the config/template file will be written regardless\n"+
				"of whether one exists or not.")
	},
}

func cmd_write(c *command) bool {
	c.assertLeastNArg(1)

	var dir string
	if arg := strings.TrimSpace(c.flags.Arg(1)); len(arg) > 0 {
		dir = arg
	} else {
		dir = strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME"))
		if len(dir) == 0 {
			dir = path.Join(os.Getenv("HOME"), ".config")
		}
		dir = path.Join(dir, "goim")
		if err := os.MkdirAll(dir, 0777); err != nil {
			pef("Could not create '%s': %s", dir, err)
			return false
		}
	}
	switch c.flags.Arg(0) {
	case "config":
		conf := []byte(strings.TrimSpace(defaultConfig) + "\n")
		return writeFile(c, path.Join(dir, "config.toml"), conf)
	case "templates":
		tpls := []byte(strings.TrimSpace(tpl.Defaults) + "\n")
		return writeFile(c, path.Join(dir, "command.tpl"), tpls)
	default:
		pef("Unknown command '%s'.", c.flags.Arg(0))
		return false
	}
}

func writeFile(c *command, fpath string, contents []byte) bool {
	if !flagConfigOverwrite {
		_, err := os.Stat(fpath)
		if !os.IsNotExist(err) {
			pef("File at '%s' already exists. Remove or use "+
				"-overwrite.", fpath)
			return false
		}
	}
	if err := ioutil.WriteFile(fpath, contents, 0666); err != nil {
		pef("Could not write '%s': %s", fpath, err)
		return false
	}
	logf("Wrote %s successfully.", fpath)
	return true
}
