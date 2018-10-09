package main

import (
	"io/ioutil"

	"github.com/golang/glog"
)

func lookupJwt() {
	buf, err := ioutil.ReadFile(*tokenPath)
	if err != nil {
		glog.Errorf("ERROR: failed to read JWT token from %s", *tokenPath)
		return
	}

	if buf == nil {
		return
	}

	s := string(buf)
	if s != "" {
		glog.Errorf("Using JWT token at %s", *tokenPath)
		return
	}

	*jwt = s
	return
}
