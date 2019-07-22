package x

import (
	"io"
	"io/ioutil"
)

func MustReadAll(r io.Reader) []byte {
	all, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	return all
}
