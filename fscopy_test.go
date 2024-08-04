package bsh

import (
	"path/filepath"
	"testing"
)

func TestCopyContents(t *testing.T) {
	b := Bsh{}

	files := []string{
		"first.txt",
		"second.bin",
		"third.jpg",
		"fourth.png",
		"fifth/sixth.foo",
		"fifth/seventh.nfo",
	}

	b.MkdirAll("local/copy_test")
	b.InDir("local/copy_test", func() {
		for _, file := range files {
			b.Touch(file)
		}
	})
	b.InDir("local", func() {
		b.RemoveAll("copy_test2")
		b.MkdirAll("copy_test2")
		b.CopyContents("copy_test", "copy_test2")
	})

	for _, path := range files {
		if !b.IsFile(filepath.Join("local/copy_test2/", path)) {
			t.Errorf("File %s does not exist", path)
		}
	}
}
