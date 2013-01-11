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
	_, err := GetPkgBuild("cower")
	if err != nil {
		t.Error(err)
	}
}

func TestTarball(t *testing.T) {
	_, err := GetTarball("cower")
	if err != nil {
		t.Error(err)
	}
}

func TestMethod(t *testing.T) {
	sr, err := GetResults("search", "pacman")
	if err != nil {
		t.Error(err)
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
