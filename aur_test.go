package main

import (
	"encoding/json"
	"log"
	"testing"
)

func init() {
	log.SetPrefix("")
	log.SetFlags(0)
}

func TestPkgbuild(t *testing.T) {
	_, err := GetPkgBuild("pacman-git")
	if err != nil {
		t.Error(err)
	}
}

func TestTarball(t *testing.T) {
	_, err := GetTarball("pacman-git")
	if err != nil {
		t.Error(err)
	}
}

func TestSearch(t *testing.T) {
	sr, err := GetResults("search", "pacman-git")
	if err != nil {
		t.Fatal(err)
	}
	for _, i := range sr.RawResults {
		b, err := i.MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}
		result := new(Result)
		err = json.Unmarshal(b, result)
		if err != nil {
			t.Fatal(err)
		}
	}
}
