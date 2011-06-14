package main

import (
	"json"
	"log"
	"testing"
	"timer"
)

func init() {
	log.SetPrefix("")
	log.SetFlags(0)
}

func TestPkgbuild(t *testing.T) {
	aur, _ := NewAur()
	defer timer.From(timer.Now())
	_, err := aur.Pkgbuild("cower")
	if err != nil {
		t.Error(err)
	}
}

func TestTarball(t *testing.T) {
	defer timer.From(timer.Now())
	aur, _ := NewAur()
	_, err := aur.Pkgbuild("cower")
	if err != nil {
		t.Error(err)
	}
}

func TestMethod(t *testing.T) {
	defer timer.From(timer.Now())
	aur, _ := NewAur()
	sr, err := aur.Results("search", "git")
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
