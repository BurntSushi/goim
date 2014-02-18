package main

import (
	"compress/gzip"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	path "path/filepath"
	"strings"
)

// fetcher provides an interface for retrieving IMDB data files.
// This abstract over where the data files come from: local directory, HTTP,
// FTP, etc.
type fetcher interface {
	list(name string) (io.ReadCloser, error)
}

// newGzipFetcher is just like newFetcher, except it's wrapped in a gzip
// reader. Use this when you intend on reading the file, and use the plain
// newFetcher when you just intend on saving to disk.
func newGzipFetcher(uri string) fetcher {
	f := newFetcher(uri)
	if f == nil {
		return nil
	}
	return gzipFetcher{f}
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
		return dirFetcher(uri)
	}

	loc, err := url.Parse(uri)
	if err != nil {
		pef("Could not parse URL '%s': %s", uri, err)
		return nil
	}
	switch loc.Scheme {
	case "http":
		return httpFetcher{loc}
	case "ftp":
		return ftpFetcher{loc}
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

type ftpReadCloser struct {
	cmd            *exec.Cmd
	stdout, stderr io.ReadCloser
}

func (r *ftpReadCloser) Read(bs []byte) (int, error) {
	n, err := r.stdout.Read(bs)
	if err != nil && err != io.EOF {
		stderr, err2 := ioutil.ReadAll(r.stderr)
		if err2 != nil {
			return 0, ef("Bad stuff happened while reading stderr: %s", err2)
		}
		return 0, ef("FTP download failed: %s\n\nstderr:\n\n%s", err, stderr)
	}
	return n, err
}

func (r *ftpReadCloser) Close() error {
	if r.cmd == nil {
		return nil
	}
	if err := r.cmd.Wait(); err != nil {
		return ef("Could not close FTP download: %s", err)
	}
	return nil
}

// ftpFetcher satisfies the fetcher interface by reading from an FTP URL.
// Each fetcher opens a new FTP connection.
type ftpFetcher struct {
	*url.URL
}

func (ff ftpFetcher) list(name string) (io.ReadCloser, error) {
	var goim string
	var err error

	if strings.Contains(os.Args[0], string(path.Separator)) {
		goim, err = path.Abs(os.Args[0])
		if err != nil {
			return nil, ef("Could not find 'goim' executable: %s", err)
		}
	} else {
		goim = "goim"
	}

	c := exec.Command(goim, "ftp", name, ff.URL.String())
	stdout, err := c.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := c.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err := c.Start(); err != nil {
		return nil, err
	}
	return &ftpReadCloser{c, stdout, stderr}, nil
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
