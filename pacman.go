package main

import (
	"path"
	"path/filepath"
	"strings"
)

const (
	localdb = "/var/lib/pacman/local"
	sycndb  = "/var/lib/pacman/sync"
)

func isInstalled(name string) bool {
	glob := path.Join(localdb, name+"-*")
	results, err := filepath.Glob(glob)
	handleError(err)
	for _, s := range results {
		s = getName(s)
		if s == name {
			return true
		}
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

func whichSync(name string) string {
	glob := path.Join(sycndb, "*/"+name+"-*")
	results, err := filepath.Glob(glob)
	handleError(err)
	for _, s := range results {
		if getName(s) == name {
			dir, _ := path.Split(s)
			return path.Base(dir)
		}
	}
	return ""
}

func whichRepo(name string) string {
	if isInstalled(name) {
		return "installed"
	}
	repo := whichSync(name)
	switch repo {
	case "":
		return "aur"
	default:
		return repo
	}
	return ""
}
