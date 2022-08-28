package bsh

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
)

// ExeName adds ".exe" to passed string if GOOS is windows
func ExeName(path string) string {
	if runtime.GOOS == "windows" {
		return path + ".exe"
	}
	return path
}

// ExeName adds ".exe" to passed string if GOOS is windows
func (b *Bsh) ExeName(path string) string {
	return ExeName(path)
}

// Getwd is os.Getwd, but with errors handled by this instance of Bsh
func (b *Bsh) Getwd() string {
	dir, err := os.Getwd()
	if err != nil {
		b.Panic(err)
	}
	return dir
}

// Chdir is os.Chdir, but with errors handled by this instance of Bsh
func (b *Bsh) Chdir(dir string) {
	b.Verbosef("Chdir: %s", dir)
	if err := os.Chdir(dir); err != nil {
		b.Panic(err)
	}
}

// MkdirAll is os.MkdirAll, but with errors handled by this instance of Bsh
func (b *Bsh) MkdirAll(dir string) {
	b.Verbosef("MkdirAll: %s", dir)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		b.Panic(err)
	}
}

// InDir saves the cwd, creates the given path (if needed), cds into the
// given path, executes the given func, then restores the previous cwd.
func (b *Bsh) InDir(path string, fn func()) {
	prev := b.Getwd()
	b.MkdirAll(path)
	b.Chdir(path)
	fn()
	b.Chdir(prev)
}

// Remove is os.Remove, but with errors handled by this instance of Bsh
func (b *Bsh) Remove(dir string) {
	b.Verbosef("Remove: %s", dir)
	if err := os.Remove(dir); err != nil {
		b.Panic(err)
	}
}

// RemoveAll is os.RemoveAll, but with errors handled by this instance of Bsh
func (b *Bsh) RemoveAll(dir string) {
	b.Verbosef("RemoveAll: %s", dir)
	if err := os.RemoveAll(dir); err != nil {
		b.Panic(err)
	}
}

// Exists checks if this path already exists on disc (as a file or folder or whatever)
func (b *Bsh) Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			b.Panic(err)
		}
		return false
	}
	return true
}

// IsFile checks if this path is a file (returns false if path doesn't exist, or exists but is a folder)
func (b *Bsh) IsFile(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			b.Panic(err)
		}
		return false
	}
	return !fi.IsDir()
}

// IsDir checks if this path is a folder (returns false if path doesn't exist, or exists but is a file)
func (b *Bsh) IsDir(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			b.Panic(err)
		}
		return false
	}
	return fi.IsDir()
}

// Stat is os.Stat, but with errors handled by this instance of Bsh
func (b *Bsh) Stat(path string) fs.FileInfo {
	b.Verbosef("Stat: %s", path)
	fi, err := os.Stat(path)
	if err != nil {
		b.Panic(err)
	}
	return fi
}

// File Copy

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

func (b *Bsh) copyImpl(src, dst string) error {
	b.Verbosef("copy: %s => %s", src, dst)
	sf, err := os.Open(src)
	if err != nil {
		if os.IsNotExist(err) {
			return err
		}
		return fmt.Errorf("error opening %s: %w", src, err)
	}
	defer sf.Close()

	info, err := sf.Stat()
	if err != nil {
		return fmt.Errorf("error reading %s: %w", src, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}
	srcSize := info.Size()

	df, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("error creating %s: %w", dst, err)
	}
	defer df.Close()

	dstSize, err := io.Copy(df, sf)
	if err != nil {
		return fmt.Errorf("error copying from %s to %s: %w", src, dst, err)
	}
	if dstSize != srcSize {
		return fmt.Errorf("%s has %d byte(s), but the copy %s only has %d byte(s)", src, srcSize, dst, dstSize)
	}
	return nil
}

// Write file (create or truncate)

func (b *Bsh) Write(path string, contents string) {
	if err := b.writeImpl(path, contents, nil, false); err != nil {
		b.Panic(err)
	}
}

func (b *Bsh) Writef(path string, format string, args ...interface{}) {
	if err := b.writeImpl(path, fmt.Sprintf(format, args...), nil, false); err != nil {
		b.Panic(err)
	}
}

func (b *Bsh) WriteErr(path string, contents string) error {
	return b.writeImpl(path, contents, nil, false)
}

func (b *Bsh) WriteBytes(path string, data []byte) {
	if err := b.writeImpl(path, "", data, false); err != nil {
		b.Panic(err)
	}
}

func (b *Bsh) WriteBytesErr(path string, data []byte) error {
	return b.writeImpl(path, "", data, false)
}

// Append file

func (b *Bsh) Append(path string, contents string) {
	if err := b.writeImpl(path, contents, nil, true); err != nil {
		b.Panic(err)
	}
}

func (b *Bsh) Appendf(path string, format string, args ...interface{}) {
	if err := b.writeImpl(path, fmt.Sprintf(format, args...), nil, true); err != nil {
		b.Panic(err)
	}
}

func (b *Bsh) AppendErr(path string, contents string) error {
	return b.writeImpl(path, contents, nil, true)
}

func (b *Bsh) AppendBytes(path string, data []byte) {
	if err := b.writeImpl(path, "", data, true); err != nil {
		b.Panic(err)
	}
}

func (b *Bsh) AppendBytesErr(path string, data []byte) error {
	return b.writeImpl(path, "", data, true)
}

func (b *Bsh) writeImpl(path string, str string, data []byte, append bool) error {
	if len(str) > 0 && len(data) > 0 {
		return fmt.Errorf("this should never happen: writeImpl has both string and []byte")
	}
	var f *os.File
	var err error
	if append {
		b.Verbosef("Append to file: %s", path)
		f, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	} else {
		b.Verbosef("Write to file: %s", path)
		f, err = os.Create(path)
	}
	if err != nil {
		return err
	}
	defer f.Close()
	if len(str) > 0 {
		_, err = io.Copy(f, strings.NewReader(str))
	} else {
		_, err = io.Copy(f, bytes.NewReader(data))
	}
	if err != nil {
		return err
	}
	return nil
}

// Read file

func (b *Bsh) Read(path string) string {
	str, err := b.ReadErr(path)
	if err != nil {
		b.Panic(err)
	}
	return str
}

func (b *Bsh) ReadErr(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	// TODO: this cast from []byte to string involves an allocation and copy.
	// Is there a way to skip that work and read straight into a string?
	return string(data), nil
}

func (b *Bsh) ReadFile(path string) []byte {
	b.Verbosef("Read from file: %s", path)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		b.Panic(err)
	}
	return data
}
