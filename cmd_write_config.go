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

func writeConfig(c *command) bool {
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
			pef("Could not create '%s': %s", dir, err)
			return false
		}
	}

	confPath := path.Join(dir, "config.toml")
	tplPath := path.Join(dir, "format.tpl")

	// Don't clobber the user's config unexpectedly!
	if !flagConfigOverwrite {
		_, err := os.Stat(confPath)
		if !os.IsNotExist(err) {
			pef("Config file at '%s' already exists. Remove or use "+
				"-overwrite.", confPath)
			return false
		}
		_, err = os.Stat(tplPath)
		if !os.IsNotExist(err) {
			pef("Template file at '%s' already exists. Remove or use "+
				"-overwrite.", tplPath)
			return false
		}
	}

	conf := []byte(strings.TrimSpace(defaultConfig) + "\n")
	if err := ioutil.WriteFile(confPath, conf, 0666); err != nil {
		pef("Could not write '%s': %s", confPath, err)
		return false
	}

	tplText := []byte(strings.TrimSpace(tpl.Defaults) + "\n")
	if err := ioutil.WriteFile(tplPath, tplText, 0666); err != nil {
		pef("Could not write '%s': %s", tplPath, err)
		return false
	}
	return true
}
