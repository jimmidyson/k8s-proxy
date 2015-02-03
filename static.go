package main

import (
	"net/http"
	"path"
	"strings"
	"time"
)

type fileHandler struct {
	httproot    http.FileSystem
	defaultPage string
}

func (f *fileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	upath := r.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
		r.URL.Path = upath
	}
	f.serveFile(w, r, f.httproot, path.Clean(upath), true)
}

// name is '/'-separated, not filepath.Separator.
func (fh *fileHandler) serveFile(w http.ResponseWriter, r *http.Request, fs http.FileSystem, name string, redirect bool) {
	const indexPage = "/index.html"

	f, err := fs.Open(name)
	if err != nil {
		if len(fh.defaultPage) > 0 {
			r.URL.Path = fh.defaultPage
			fh.ServeHTTP(w, r)
		} else {
			http.NotFound(w, r)
		}
		return
	}
	defer f.Close()

	d, err1 := f.Stat()
	if err1 != nil {
		if len(fh.defaultPage) > 0 {
			r.URL.Path = fh.defaultPage
			fh.ServeHTTP(w, r)
		} else {
			http.NotFound(w, r)
		}
		return
	}

	if redirect {
		// redirect to canonical path: / at end of directory url
		// r.URL.Path always begins with /
		url := r.URL.Path
		if d.IsDir() {
			if url[len(url)-1] != '/' {
				localRedirect(w, r, path.Base(url)+"/")
				return
			}
		} else {
			if url[len(url)-1] == '/' {
				localRedirect(w, r, "../"+path.Base(url))
				return
			}
		}
	}

	// use contents of index.html for directory, if present
	if d.IsDir() {
		index := strings.TrimSuffix(name, "/") + indexPage
		ff, err := fs.Open(index)
		if err == nil {
			defer ff.Close()
			dd, err := ff.Stat()
			if err == nil {
				name = index
				d = dd
				f = ff
			}
		}
	}

	// Still a directory? (we didn't find an index.html file)
	if d.IsDir() {
		if checkLastModified(w, r, d.ModTime()) {
			return
		}
		http.NotFound(w, r)
		return
	}

	http.ServeContent(w, r, d.Name(), d.ModTime(), f)
}

func checkLastModified(w http.ResponseWriter, r *http.Request, modtime time.Time) bool {
	if modtime.IsZero() {
		return false
	}

	// The Date-Modified header truncates sub-second precision, so
	// use mtime < t+1s instead of mtime <= t to check for unmodified.
	if t, err := time.Parse(http.TimeFormat, r.Header.Get("If-Modified-Since")); err == nil && modtime.Before(t.Add(1*time.Second)) {
		h := w.Header()
		delete(h, "Content-Type")
		delete(h, "Content-Length")
		w.WriteHeader(http.StatusNotModified)
		return true
	}
	w.Header().Set("Last-Modified", modtime.UTC().Format(http.TimeFormat))
	return false
}

func localRedirect(w http.ResponseWriter, r *http.Request, newPath string) {
	if q := r.URL.RawQuery; q != "" {
		newPath += "?" + q
	}
	w.Header().Set("Location", newPath)
	w.WriteHeader(http.StatusMovedPermanently)
}

func FileServerWithDefault(root http.FileSystem, defaultPage string) http.Handler {
	return &fileHandler{root, defaultPage}
}
