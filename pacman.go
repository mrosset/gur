package main

import (
	"archive/tar"
	"bytes"
	"path"
	"path/filepath"
	"strings"
	"compress/gzip"
	"io"
	"os"
)

var (
	packages  = map[string]map[string]string{}
	installed = map[string]bool{}
)

const (
	localdb = "/var/lib/pacman/local"
	sycndb  = "/var/lib/pacman/sync"
)


func isInstalled(name string) bool {
	installed, ok := installed[name]
	if installed && ok {
		return true
	}
	return false
}

func getName(p string) string {
	p = path.Base(p)
	v := strings.Split(p, "-", -1)
	// remove version and package version
	v = v[0 : len(v)-2]
	return strings.Join(v, "-")
}

func readInstalled() {
	glob := path.Join(localdb, "*")
	results, err := filepath.Glob(glob)
	handleError(err)
	for _, s := range results {
		s = getName(s)
		installed[s] = true
	}
}

func readSyncDB(p string, c chan int) {
	repo := path.Base(path.Base(p))
	repo = repo[0 : len(repo)-3]
	f, err := os.Open(p)
	handleError(err)
	defer f.Close()
	gr, err := gzip.NewReader(f)
	handleError(err)
	tr := tar.NewReader(gr)
	for {
		h, err := tr.Next()
		if err != nil && err != os.EOF {
			handleError(err)
		}
		if h == nil || err == os.EOF {
			break
		}
		if h.Typeflag == tar.TypeDir {
			buf := new(bytes.Buffer)
			tr.Next()
			io.Copy(buf, tr)
			tr.Next()
			io.Copy(buf, tr)
			parseMeta(buf, repo)
		}
	}
	c <- 1
}

func parseMeta(buf *bytes.Buffer, repo string) {
	var name string
	pack := map[string]string{}
	pack["REPO"] = repo
	_, _ = buf.ReadByte()
	for {
		key, err := buf.ReadBytes('%')
		if err == os.EOF {
			break
		}
		key = bytes.Trim(key, "%")
		values, _ := buf.ReadBytes('%')
		values = bytes.Trim(values, "%")
		values = bytes.Replace(values, []byte("\n"), []byte(" "), -1)
		values = bytes.Trim(values, " ")
		if string(key) == "NAME" {
			name = string(values)
		}
		pack[string(key)] = string(values)
	}
	v, _ := packages[name]
	if v != nil {
		printf("%s exists\n", v)
	}
	packages[name] = pack
}

func loadSyncCache() {
	glob := path.Join(sycndb, "*.db")
	results, err := filepath.Glob(glob)
	handleError(err)
	c := make(chan int)
	for _, v := range results {
		printf("reading %s\n", v)
		go readSyncDB(v, c)
	}
	for _ = range results {
		<-c
	}
}

func whichRepo(pack string) (string, string) {
	if isInstalled(pack) {
		return "installed", ""
	}
	_, ok := packages[pack]
	if ok {
		return packages[pack]["REPO"], ""
	}
	p := findProvides(pack)
	if p == nil {
		return "aur", ""
	}
	return p["REPO"], p["NAME"]
}

func findProvides(pack string) map[string]string {
	for _, p := range packages {
		if strings.Contains(p["PROVIDES"], pack) {
			return p
		}
	}
	return nil
}
