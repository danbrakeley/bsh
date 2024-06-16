package bsh

import (
	"os"
	"testing"
)

func ensureLocalFolder(t *testing.T) {
	t.Helper()
	err := os.MkdirAll("local", os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
}

func TestZipFile(t *testing.T) {
	ensureLocalFolder(t)
	b := Bsh{}
	b.InDir("local", func() {
		b.RemoveAll("zip_test.zip")
		b.Touch("zip_test.txt")
		b.ZipFile("zip_test.txt", "zip_test.zip")
		if !b.Exists("zip_test.zip") {
			t.Fatal("ZipExe did not produce output")
		}
		// TODO: when unzip is added, use that to test the zip contents here
	})
}

func TestZipFolder(t *testing.T) {
	ensureLocalFolder(t)
	b := Bsh{}
	b.InDir("local", func() {
		b.RemoveAll("zip_test.zip")
		b.Touch("zip_test.txt")
		b.Touch("foo/test2")
		b.Touch("foo/test3")
		b.ZipFolder(".", "zip_test.zip")
		if !b.Exists("zip_test.zip") {
			t.Fatal("ZipExe did not produce output")
		}
		// TODO: when unzip is added, use that to test the zip contents here
	})
}
