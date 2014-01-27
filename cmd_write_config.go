package main

import (
	"flag"
	"io/ioutil"
	"os"
	"path"
	"strings"

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

var xdgPaths = xdg.Paths{XDGSuffix: "goim"}

var cmdWriteConfig = &command{
	name:            "write-config",
	positionalUsage: "[ file ]",
	shortHelp:       "write a default configuration",
	help: `
Writes the default configuration to $XDG_CONFIG_HOME/goim/config.toml or to
the file argument given.

If no argument is given and $XDG_CONFIG_HOME is not set, then the configuration
is written to $HOME/.config/goim/config.toml.
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
	var fpath string
	if arg := strings.TrimSpace(c.flags.Arg(0)); len(arg) > 0 {
		fpath = arg
	} else {
		pxdg := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME"))
		if len(pxdg) == 0 {
			pxdg = path.Join(os.Getenv("HOME"), ".config")
		}
		pxdg = path.Join(pxdg, "goim")
		if err := os.MkdirAll(pxdg, 0777); err != nil {
			fatalf("Could not create '%s': %s", pxdg, err)
		}
		fpath = path.Join(pxdg, "config.toml")
	}

	// Don't clobber the user's config unexpectedly!
	if !flagConfigOverwrite {
		_, err := os.Stat(fpath)
		if !os.IsNotExist(err) {
			fatalf("Config file at '%s' already exists. Remove or use "+
				"-overwrite.", fpath)
		}
	}

	conf := []byte(strings.TrimSpace(defaultConfig) + "\n")
	if err := ioutil.WriteFile(fpath, conf, 0666); err != nil {
		fatalf("Could not write '%s': %s", fpath, err)
	}
}
