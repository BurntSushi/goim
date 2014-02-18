package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
	"text/tabwriter"

	"github.com/BurntSushi/ty/fun"
)

var commands = []*command{
	cmdFull,
	cmdShort,
	cmdLoad,
	cmdSearch,
	cmdSize,
	cmdWriteConfig,
	cmdRename,
	cmdFtp,
}

func usage() {
	pef("goim is a tool for interacting with a local copy of IMDB.\n")
	pef("Usage:\n\n    goim {command} [flags] [arguments]\n")
	pef("Use 'goim help {command}' for more details on {command}.\n")

	fun.Sort(func(c1, c2 *command) bool { return c1.name < c2.name }, commands)

	pef("A list of the main commands:\n")
	tabw := tabwriter.NewWriter(os.Stderr, 0, 0, 4, ' ', 0)
	for _, c := range commands {
		if c.name == "ftp" || c.other {
			continue
		}
		fmt.Fprintf(tabw, "    %s\t%s\n", c.name, c.shortHelp)
	}
	tabw.Flush()
	pef("")

	pef("A list of other commands:\n")
	for _, c := range commands {
		if c.name == "ftp" || !c.other {
			continue
		}
		fmt.Fprintf(tabw, "    %s\t%s\n", c.name, c.shortHelp)
	}
	tabw.Flush()
	pef("")
	os.Exit(1)
}

func main() {
	var cmd string
	var help bool
	if len(os.Args) < 2 {
		usage()
	} else if strings.TrimLeft(os.Args[1], "-") == "help" {
		if len(os.Args) < 3 {
			usage()
		} else {
			cmd = os.Args[2]
			help = true
		}
	} else {
		cmd = os.Args[1]
	}

	for _, c := range commands {
		if c.name == cmd {
			c.setCommonFlags()
			if c.addFlags != nil {
				c.addFlags(c)
			}
			if help {
				c.showHelp()
			} else {
				c.flags.Usage = c.showUsage
				c.flags.Parse(os.Args[2:])

				if flagCpu < 1 {
					flagCpu = 1
				}
				runtime.GOMAXPROCS(flagCpu)

				if len(flagCpuProfile) > 0 {
					f := createFile(flagCpuProfile)
					pprof.StartCPUProfile(f)
					defer f.Close()
					defer pprof.StopCPUProfile()
				}

				if !c.run(c) {
					os.Exit(1)
				}
				return
			}
		}
	}
	pef("Unknown command '%s'. Run 'goim help' for a list of "+
		"available commands.", cmd)
	os.Exit(1)
}
