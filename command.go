package main

import (
	"flag"
	"fmt"

	"github.com/BurntSushi/goim/imdb"

	"log"
	"os"
	"runtime"
	"strings"
)

var (
	flagCpuProfile = ""
	flagCpu        = runtime.NumCPU()
	flagQuiet      = false
	flagDb         = ""
)

func init() {
	log.SetFlags(0)
}

type command struct {
	name            string
	positionalUsage string
	shortHelp       string
	help            string
	flags           *flag.FlagSet
	addFlags        func(*command)
	run             func(*command)
}

func (c *command) showUsage() {
	log.Printf("Usage: goim %s [flags] %s\n", c.name, c.positionalUsage)
	c.showFlags()
	os.Exit(1)
}

func (c *command) showHelp() {
	log.Printf("Usage: goim %s [flags] %s\n\n", c.name, c.positionalUsage)
	log.Println(strings.TrimSpace(c.help))
	log.Printf("\nThe flags are:\n\n")
	c.showFlags()
	log.Println("")
	os.Exit(1)
}

func (c *command) showFlags() {
	c.flags.VisitAll(func(fl *flag.Flag) {
		var def string
		if len(fl.DefValue) > 0 {
			def = fmt.Sprintf(" (default: %s)", fl.DefValue)
		}
		usage := strings.Replace(fl.Usage, "\n", "\n    ", -1)
		log.Printf("-%s%s\n", fl.Name, def)
		log.Printf("    %s\n", usage)
	})
}

func (c *command) setCommonFlags() {
	c.flags.StringVar(&flagDb, "db", flagDb,
		"Overrides the database to be used. It should be a string of the "+
			"form 'driver:dsn'.")
	c.flags.StringVar(&flagCpuProfile, "cpu-prof", flagCpuProfile,
		"When set, a CPU profile will be written to the file path provided.")
	c.flags.IntVar(&flagCpu, "cpu", flagCpu,
		"Sets the maximum number of CPUs that can be executing simultaneously.")
	c.flags.BoolVar(&flagQuiet, "quiet", flagQuiet,
		"When set, status messages about the progress of a command will be "+
			"omitted.")
}

func (c *command) db() *imdb.DB {
	if len(flagDb) == 0 {
		fatalf("Configuration not yet supported.")
	}
	dbInfo := strings.Split(flagDb, ":")
	driver, dsn := dbInfo[0], dbInfo[1]
	db, err := imdb.Open(driver, dsn)
	if err != nil {
		fatalf("Could not open '%s': %s", flagDb, err)
	}
	return db
}

func (c *command) assertNArg(n int) {
	if c.flags.NArg() != n {
		c.showUsage()
	}
}

func (c *command) assertLeastNArg(n int) {
	if c.flags.NArg() < n {
		c.showUsage()
	}
}
