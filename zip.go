package bsh

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func (b *Bsh) ZipFile(source, target string) {
	b.Verbosef("ZipFile: %s to %s", source, target)
	if err := zipFile(source, target, nil); err != nil {
		b.Panic(err)
	}
}

func (b *Bsh) ZipFileMode(source, target string, mode fs.FileMode) {
	b.Verbosef("ZipFileMode: %s with mode 0o%o to %s", source, mode, target)
	if err := zipFile(source, target, &mode); err != nil {
		b.Panic(err)
	}
}

func (b *Bsh) ZipFolder(source, target string) {
	b.Verbosef("ZipFolder: %s to %s", source, target)
	if err := zipFolder(source, target); err != nil {
		b.Panic(err)
	}
}

func zipFile(source, target string, mode *fs.FileMode) error {
	fzip, err := os.Create(target)
	if err != nil {
		return err
	}
	defer fzip.Close()

	zw := zip.NewWriter(fzip)
	defer zw.Close()

	info, err := os.Stat(source)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return fmt.Errorf("%s is a directory", source)
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Method = zip.Deflate
	if mode != nil {
		header.SetMode(*mode)
	}

	hw, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}

	fsrc, err := os.Open(source)
	if err != nil {
		return err
	}
	defer fsrc.Close()

	_, err = io.Copy(hw, fsrc)
	return err
}

func zipFolder(source, target string) error {
	files := make([]string, 0, 256)

	err := fs.WalkDir(os.DirFS(source), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return err
	}

	fzip, err := os.Create(target)
	if err != nil {
		return err
	}
	defer fzip.Close()

	zw := zip.NewWriter(fzip)
	defer zw.Close()

	for _, file := range files {
		path := filepath.Join(source, file)
		info, err := os.Stat(path)
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Method = zip.Deflate
		header.Name = file

		if info.IsDir() {
			header.Name += "/"
		}

		hw, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			continue
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(hw, f)
		if err != nil {
			return err
		}
	}

	return nil
}
