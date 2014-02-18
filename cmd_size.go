package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/BurntSushi/csql"

	"github.com/BurntSushi/goim/imdb"
)

var cmdSize = &command{
	name:      "size",
	shortHelp: "lists size of tables and total size of database",
	help:      "",
	flags:     flag.NewFlagSet("size", flag.ExitOnError),
	run:       cmd_size,
}

func cmd_size(c *command) bool {
	db := openDb(c.dbinfo())
	defer closeDb(db)

	tables, err := db.Tables()
	if err != nil {
		pef("%s", err)
		return false
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 2, 4, ' ', 0)
	for _, table := range tables {
		fmt.Fprintf(tw, "%s\t%s\n", table, tableSize(db, table))
	}

	_, dsn := c.dbinfo()
	total := databaseSize(db, dsn)
	fmt.Fprintf(tw, "total\t%s\n", total)
	tw.Flush()
	return true
}

// tableSize returns a pretty string indicating the size in table. Row count
// is always include, but the size on disk is only included if it's supported
// by the database.
// Note that 'name' is assumed to be SQL-safe.
func tableSize(db *imdb.DB, name string) string {
	count := csql.Count(db, sf("SELECT COUNT(*) AS count FROM %s", name))
	if db.Driver == "sqlite3" {
		return sf("%d rows", count)
	}
	var size string
	q := sf("SELECT pg_size_pretty(pg_relation_size('%s'))", name)
	csql.Scan(db.QueryRow(q), &size)
	return sf("%d rows (%s)", count, size)
}

// databaseSize returns a pretty string indicating the size of the entire
// database on disk.
func databaseSize(db *imdb.DB, dsn string) string {
	if db.Driver == "sqlite3" {
		fi, err := os.Stat(dsn)
		csql.Panic(err)
		return prettyFileSize(fi.Size())
	}
	var size string
	q := sf("SELECT pg_size_pretty(pg_database_size(current_database()))")
	csql.Scan(db.QueryRow(q), &size)
	return size
}

func prettyFileSize(bytes int64) string {
	cutoff := int64(1024 * 2)
	kb, mb, gb := int64(1024), int64(1024*1024), int64(1024*1024*1024)
	if bytes < cutoff {
		return sf("%d bytes", bytes)
	}
	kbytes := bytes / kb
	if kbytes < cutoff {
		return sf("%d kB", kbytes)
	}
	mbytes := bytes / mb
	if mbytes < cutoff {
		return sf("%d MB", mbytes)
	}
	return sf("%d GB", bytes/gb)
}
