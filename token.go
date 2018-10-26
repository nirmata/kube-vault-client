package main

import (
	"io/ioutil"
)

func lookupJwt() (string, error) {
	buf, err := ioutil.ReadFile(*tokenPath)
	if err != nil {
		return "", err
	}

	s := string(buf)
	return s, nil
}
