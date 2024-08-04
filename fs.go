package bsh

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

// Touch creates a file if it doesn't exist, and creates any intermediate folders needed.
func (b *Bsh) Touch(path string) {
	b.Verbosef("Touch: %s", path)

	dir := filepath.Dir(path)
	if len(dir) > 0 && dir != "." && dir != "/" && dir != "\\" {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			b.Panic(err)
		}
	}

	f, err := os.Create(path)
	if err != nil {
		b.Panic(err)
	}
	f.Close()
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

// InDir saves the cwd, creates the given path (if needed), cds into the
// given path, executes the given func, then restores the previous cwd.
func (b *Bsh) InDir(path string, fn func()) {
	// no need to verbose anything here, as MkdirAll and/or Chdir will verbose for us
	prev := b.Getwd()
	if !b.Exists(path) {
		b.MkdirAll(path)
	}
	b.Chdir(path)
	defer b.Chdir(prev)
	fn()
}

// MkdirTemp creates a new temp folder and returns its path and a cleanup function.
// The cleanup function deletes the temp folder and all its contents.
func (b *Bsh) MkdirTemp() (tmpdir string, cleanup func()) {
	ostmp := os.TempDir()
	tmpdir, err := os.MkdirTemp(ostmp, "bsh_*")
	if err != nil {
		b.Panic(err)
	}
	b.Verbosef("MkdirTemp: %s", tmpdir)
	return tmpdir, func() {
		b.RemoveAll(tmpdir)
	}
}

// InTempDir is like InDir, but uses a unique and newly created temp folder
// instead of a passed folder name.
// The temp folder is deleted before this func returns.
func (b *Bsh) InTempDir(fn func()) {
	tmpdir, cleanup := b.MkdirTemp()
	defer cleanup()
	b.InDir(tmpdir, fn)
}

// IsExeInPath calls exec.LookPath to locate an executable in the PATH environment var.
// Returns true if the given executable is found, otherwise false.
func (b *Bsh) IsExeInPath(file string) bool {
	path, err := exec.LookPath(file)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return false
		}
		b.Panic(err)
	}
	return len(path) > 0
}
