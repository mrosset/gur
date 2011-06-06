package aur

import (
	"testing"
)

func TestPkgbuild(t *testing.T) {
	aur, _ := NewAur()
	_, err := aur.Pkgbuild("cower")
	if err != nil {
		t.Error(err)
	}
}

func TestTarball(t *testing.T) {
	aur, _ := NewAur()
	_, err := aur.Pkgbuild("cower")
	if err != nil {
		t.Error(err)
	}
}

func BenchmarkPkgbuild(b *testing.B) {
	aur, _ := NewAur()
	for i := 0; i < b.N; i++ {
		_, _ = aur.Pkgbuild("cower")
	}
}

func BenchmarkTarball(b *testing.B) {
	aur, _ := NewAur()
	for i := 0; i < b.N; i++ {
		_, _ = aur.Tarball("cower")
	}
}

/*
func TestRawString(b *testing.B) {
        null := bytes.NewBuffer(nil)
        for i := 0; i < b.N; i++ {
                null.WriteString(longtext)
                null.Truncate(0)
        }
}
*/
