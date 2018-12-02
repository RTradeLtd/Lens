package mocks

import (
	"errors"
	"io/ioutil"
	"os"
)

// StubIpfsCat returns a stub function for use in testing. It errors on blank
// paths, otherwise it opens file at given path and returns its contents
func StubIpfsCat(assetPath string) func(h string) (c []byte, e error) {
	return func(h string) (c []byte, e error) {
		if h != "" {
			f, err := os.Open(assetPath)
			if err != nil {
				return nil, err
			}
			defer f.Close()
			return ioutil.ReadAll(f)
		}
		return nil, errors.New("oh no")
	}
}
