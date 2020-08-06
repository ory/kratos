package pkgerx

import (
	"io"
	"io/ioutil"
)

func MustRead(closer io.ReadCloser, err error) []byte {
	out, err := ioutil.ReadAll(closer)
	if err != nil {
		_ = closer.Close()
		panic(err)
	}

	_ = closer.Close()
	return out
}
