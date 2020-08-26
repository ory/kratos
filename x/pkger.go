package x

import (
	"io/ioutil"

	"github.com/markbates/pkger/pkging"
)

func MustPkgerRead(f pkging.File, err error) []byte {
	if err != nil {
		panic(err)
	}
	defer f.Close()
	return MustReadAll(f)
}

func PkgerRead(f pkging.File, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}
