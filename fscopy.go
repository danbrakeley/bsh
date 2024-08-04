package bsh

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Copy attempts to open file at src and create/overwrite new file at dst, then copy the contents.
// If src does not exist, Copy returns false, otherwise it returns true. Other errors will panic.
func (b *Bsh) Copy(src, dst string) bool {
	err := b.copyImpl(src, dst)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		b.Panic(err)
	}
	return true
}

// MustCopy attempts to open file at src and create/overwrite new file at dst, then copy the contents.
// Any error in this process will panic.
func (b *Bsh) MustCopy(src, dst string) {
	err := b.copyImpl(src, dst)
	if err != nil {
		b.Panic(err)
	}
}

// CopyContents finds all files/folders contained in src, and then copies them into dst,
// in that order. This ensures copying into a subfolder of src doesn't recurse forever.
// Src and dst must both exist and be folders. Duplicates in dst will be overwritten.
func (b *Bsh) CopyContents(src, dst string) {
	if !b.IsDir(src) {
		b.Panic(fmt.Errorf("src %s is not a folder or does not exist", src))
	}
	if !b.IsDir(dst) {
		b.Panic(fmt.Errorf("dst %s is not a folder or does not exist", dst))
	}

	toCopy := make([]copyEntry, 0, 1024)
	toCopy = b.buildCopyList(src, dst, toCopy)
	for _, entry := range toCopy {
		if entry.isDir {
			b.MkdirAll(entry.dstPath)
		} else {
			b.MustCopy(entry.srcPath, entry.dstPath)
		}
	}
}

type copyEntry struct {
	srcPath string
	dstPath string
	isDir   bool
}

func (b *Bsh) buildCopyList(src, dst string, files []copyEntry) []copyEntry {
	contents, err := os.ReadDir(src)
	if err != nil {
		b.Panic(err)
	}
	for _, entry := range contents {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		files = append(files, copyEntry{srcPath, dstPath, entry.IsDir()})
		if entry.IsDir() {
			files = b.buildCopyList(srcPath, dstPath, files)
		}
	}
	return files
}

func (b *Bsh) copyImpl(src, dst string) error {
	b.Verbosef("Copy: %s => %s", src, dst)
	sf, err := os.Open(src)
	if err != nil {
		if os.IsNotExist(err) {
			return err
		}
		return fmt.Errorf("error opening src %s: %w", src, err)
	}
	defer sf.Close()

	info, err := sf.Stat()
	if err != nil {
		return fmt.Errorf("error reading src %s: %w", src, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}
	srcSize := info.Size()

	df, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("error creating dst %s: %w", dst, err)
	}
	defer df.Close()

	dstSize, err := io.Copy(df, sf)
	if err != nil {
		return fmt.Errorf("error copying from src %s to dst %s: %w", src, dst, err)
	}
	if dstSize != srcSize {
		return fmt.Errorf("%s has %d byte(s), but the copy %s only has %d byte(s)", src, srcSize, dst, dstSize)
	}
	return nil
}
