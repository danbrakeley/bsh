package bsh

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

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
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	// TODO: this cast from []byte to string involves an allocation and copy.
	// Is there a way to skip that work and read straight into a string?
	return string(data), nil
}

func (b *Bsh) ReadFile(path string) []byte {
	b.Verbosef("Read from file: %s", path)
	data, err := os.ReadFile(path)
	if err != nil {
		b.Panic(err)
	}
	return data
}
