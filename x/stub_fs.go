package x

import (
	"io"
	"io/fs"
	"time"
)

type StubFS struct {
	name   string
	data   []byte
	offset int
}

func NewStubFS(name string, data []byte) fs.FS {
	return &StubFS{
		name: name,
		data: data,
	}
}

func (stub *StubFS) Mode() fs.FileMode {
	return 0444
}

func (stub *StubFS) ModTime() time.Time {
	return time.Time{}
}

func (stub *StubFS) IsDir() bool {
	return false
}

func (stub *StubFS) Sys() interface{} {
	return nil
}

func (stub *StubFS) Stat() (fs.FileInfo, error) {
	return stub, nil
}

func (stub *StubFS) Read(bytes []byte) (int, error) {
	if stub.offset >= len(stub.data) {
		return 0, io.EOF
	}
	n := copy(bytes, stub.data[stub.offset:])
	stub.offset += n
	return n, nil
}

func (stub *StubFS) Close() error {
	return nil
}

func (stub *StubFS) Open(name string) (fs.File, error) {
	return stub, nil
}

func (stub *StubFS) Name() string {
	return stub.name
}

func (stub *StubFS) Size() int64 {
	return int64(len(stub.data))
}
