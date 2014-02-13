package main

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"

	ftp "github.com/jlaffaye/goftp"
)

// fetcher provides an interface for retrieving IMDB data files.
// This abstract over where the data files come from: local directory, HTTP,
// FTP, etc.
type fetcher interface {
	list(name string) (io.ReadCloser, error)
}

// newFetcher returns a fetcher based on the uri given. The uri may be a
// preset FTP site ("berlin", "digital", "funet" or "uiuc"), a full FTP or
// HTTP URL containing IMDB's list files, or a local directory containing
// IMDB's list files.
func newFetcher(uri string) fetcher {
	if v, ok := namedFtp[uri]; ok {
		uri = v
	}
	if !strings.HasPrefix(uri, "http") && !strings.HasPrefix(uri, "ftp") {
		return gzipFetcher{dirFetcher(uri)}
	}

	loc, err := url.Parse(uri)
	if err != nil {
		pef("Could not parse URL '%s': %s", uri, err)
		return nil
	}
	switch loc.Scheme {
	case "http":
		return gzipFetcher{httpFetcher{loc}}
	case "ftp":
		// It seems like FTP sites---even if they are public---require a
		// trivial login.
		if loc.User == nil {
			loc.User = url.UserPassword("anonymous", "anonymous")
		}
		// The FTP package I'm using requires a port number.
		if !strings.Contains(loc.Host, ":") {
			loc.Host += ":21"
		}

		// We're only allowed a limited number of FTP connections, so start
		// a pool of them.
		pool := ftpPool(*loc)
		return gzipFetcher{ftpFetcher{loc, pool}}
	}
	pef("Unsupported URL scheme '%s' in '%s'.", loc.Scheme, uri)
	return nil
}

// dirFetcher satisfies the fetcher interface by reading from a local
// directory.
type dirFetcher string

func (df dirFetcher) list(name string) (io.ReadCloser, error) {
	fpath := path.Join(string(df), sf("%s.list.gz", name))
	f, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// httpFetcher satisfies the fetcher interface by reading from an HTTP URL.
type httpFetcher struct {
	*url.URL
}

func (hf httpFetcher) list(name string) (io.ReadCloser, error) {
	uri := sf("%s/%s.list.gz", hf.String(), name)
	resp, err := http.Get(uri)
	if err != nil {
		return nil, ef("Could not download '%s': %s", uri, err)
	}
	return resp.Body, nil
}

const (
	maxFtpConns = 5
)

var (
	pools     = make(map[url.URL]chan *ftp.ServerConn)
	poolsLock = new(sync.Mutex)
)

func ftpPool(loc url.URL) chan *ftp.ServerConn {
	poolsLock.Lock()
	defer poolsLock.Unlock()

	if pool, ok := pools[loc]; ok {
		return pool
	}
	connChan := make(chan *ftp.ServerConn)
	pools[loc] = connChan

	c, err := newFtpConn(loc)
	if err != nil {
		fatalf("%s", err)
	}
	conns := []*ftp.ServerConn{c}
	numGiven := 0
	go func() {
		for {
			if len(conns) == 0 {
				if numGiven >= maxFtpConns {
					conns = append(conns, <-connChan)
					numGiven--
				} else {
					c, err := newFtpConn(loc)
					if err != nil {
						pef("%s", err)
						// the client will see the nil conn.
						// they can decide what to do.
						conns = append(conns, nil)
					} else {
						conns = append(conns, c)
					}
				}
			}
			select {
			case c := <-connChan:
				conns = append(conns, c)
				numGiven--
			case connChan <- conns[0]:
				conns = conns[1:]
				numGiven++
			}
		}
	}()
	return connChan
}

func newFtpConn(loc url.URL) (*ftp.ServerConn, error) {
	conn, err := ftp.Connect(loc.Host)
	if err != nil {
		return nil, ef("Could not connect to '%s': %s", loc.Host, err)
	}

	pass, _ := loc.User.Password()
	if err := conn.Login(loc.User.Username(), pass); err != nil {
		return nil, ef("Authentication failed for '%s': %s", loc.Host, err)
	}
	return conn, nil
}

// ftpFetcher satisfies the fetcher interface by reading from an FTP URL.
// Each fetcher opens a new FTP connection.
type ftpFetcher struct {
	*url.URL
	pool chan *ftp.ServerConn
}

// ftpRetrCloser syncs the closing of the file download with the closing of
// the connection.
type ftpRetrCloser struct {
	io.ReadCloser
	pool chan *ftp.ServerConn
	conn *ftp.ServerConn
}

// Close closes the file download and the FTP connection.
func (r *ftpRetrCloser) Close() error {
	defer func() {
		r.ReadCloser = nil
		r.pool <- r.conn // done with the connection
	}()

	if r.ReadCloser == nil {
		return nil
	}
	if err := r.ReadCloser.Close(); err != nil {
		pef("Problem closing FTP reader: %s", err)
		return ef("Problem closing FTP reader: %s", err)
	}
	return nil
}

func (ff ftpFetcher) list(name string) (io.ReadCloser, error) {
	conn := <-ff.pool
	if conn == nil {
		return nil, ef("Could not get FTP connection from pool.")
	}
	namePath := sf("%s/%s.list.gz", ff.Path, name)
	r, err := conn.Retr(namePath)
	if err != nil {
		return nil, ef("Could not retrieve '%s' from '%s': %s",
			namePath, ff.Host, err)
	}
	return &ftpRetrCloser{r, ff.pool, conn}, nil
}

// gzipFetcher wraps a value satisfying the fetcher interface with a gzip
// reader. It also couples the closing of a gzip reader with closing the
// underlying reader.
type gzipFetcher struct {
	fetcher
}

func (gf gzipFetcher) list(name string) (io.ReadCloser, error) {
	plain, err := gf.fetcher.list(name)
	if err != nil {
		return nil, err
	}

	gzlist, err := gzip.NewReader(plain)
	if err != nil {
		return nil, ef("Could not create gzip reader for '%s': %s", name, err)
	}
	return &gzipCloser{gzlist, plain}, nil
}

type gzipCloser struct {
	*gzip.Reader
	underlying io.ReadCloser
}

func (gc *gzipCloser) Close() error {
	defer func() {
		gc.Reader = nil
		gc.underlying = nil
	}()

	// It's important not to try closing more than once, particularly for
	// readers the originate from an FTP connection.
	if gc.Reader == nil || gc.underlying == nil {
		return nil
	}

	var err error
	if err = gc.Reader.Close(); err != nil {
		pef("Error closing gzip reader: %s", err)
	}
	if err = gc.underlying.Close(); err != nil {
		pef("Error closing initial source: %s", err)
	}
	return err
}
