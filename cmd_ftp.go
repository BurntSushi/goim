package main

import (
	"flag"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/jlaffaye/ftp"
)

const maxFtpConns = 5

var cmdFtp = &command{
	name: "ftp",
	positionalUsage: "list-name ( berlin | digital | funet | uiuc | " +
		"ftp://... )",
	shortHelp: "downloads list files from FTP server",
	help: `
This is an undocumented command that downloads the given list from the given
FTP server and prints the contents of the file to stdout. The contents are
compressed in the gzip format.

The purpose of this command is to isolate each FTP connection in a single
process. I've had major problems closing/killing an FTP connection after a long
download, so I've chosen to follow Ken Thompson's advice and use brute force.

Here's more of the story: http://goo.gl/43ICUs
`,
	flags: flag.NewFlagSet("ftp", flag.ExitOnError),
	run:   cmd_ftp,
}

var namedFtp = map[string]string{
	"berlin":  "ftp://ftp.fu-berlin.de/pub/misc/movies/database",
	"digital": "ftp://gatekeeper.digital.com.au/pub/imdb",
	"funet":   "ftp://ftp.funet.fi/pub/culture/tv+film/database",
	"uiuc":    "ftp://uiarchive.cso.uiuc.edu/pub/info/imdb",
}

func cmd_ftp(c *command) bool {
	c.assertLeastNArg(2)

	listName, uri := c.flags.Arg(0), c.flags.Arg(1)
	if v, ok := namedFtp[uri]; ok {
		uri = v
	}
	loc, err := url.Parse(uri)
	if err != nil {
		pef("Could not parse URL '%s': %s", uri, err)
		return false
	}
	if loc.User == nil {
		loc.User = url.UserPassword("anonymous", "anonymous")
	}
	if !strings.Contains(loc.Host, ":") {
		loc.Host += ":21"
	}

	conn, err := ftp.Connect(loc.Host)
	if err != nil {
		pef("Could not connect to '%s': %s", loc.Host, err)
		return false
	}

	pass, _ := loc.User.Password()
	if err := conn.Login(loc.User.Username(), pass); err != nil {
		pef("Authentication failed for '%s': %s", loc.Host, err)
		return false
	}

	namePath := sf("%s/%s.list.gz", loc.Path, listName)
	r, err := conn.Retr(namePath)
	if err != nil {
		pef("Could not retrieve '%s' from '%s': %s", namePath, loc.Host, err)
		return false
	}

	if _, err := io.Copy(os.Stdout, r); err != nil {
		pef("Could not write '%s' to stdout: %s", listName, err)
		return false
	}

	// Don't even bother trying to close the connection.
	return true
}
