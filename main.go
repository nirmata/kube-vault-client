package main

import (
	"flag"
	"strconv"

	"github.com/golang/glog"
)

var url = flag.String("url", "http://127.0.0.1:8200", "the Vault server URL")
var kubeAuthPath = flag.String("kubePath", "kubernetes", "the Vault path for Kubernetes auth (e.g. kubernetes/prod)")
var role = flag.String("role", "", "the role to use for Vault authentication")
var jwt = flag.String("jwt", "", "the token to use for Vault authentication")
var secrets = flag.String("secrets", "", "a comma separated list of paths, keys, and variable names e.g (/secret/s1#k1#name, /secret/s1#k2#name, /secret/s2#k5#name")
var tokenPath = flag.String("tokenPath", "/var/run/secrets/kubernetes.io/serviceaccount/token", "location of token - used if a token is not provided.")
var out = flag.String("out", "/var/vault/secrets", "location to store the secrets fetched from Vault")

func main() {

	flag.Parse()

	flag.Set("logtostderr", "true")
	if f := flag.CommandLine.Lookup("v"); f != nil {
		s := f.Value.String()
		l, _ := strconv.Atoi(s)
		if l < 3 {
			flag.Set("v", "3")
		}
	}

	if *url == "" || *secrets == "" || *role == "" {
		glog.Infof("Usage: ")
		flag.Usage()
		return
	}

	if *jwt == "" {
		lookupJwt()
		if *jwt == "" {
			glog.Errorf("ERROR: failed to retrieve JWT")
			return
		}
	}

	client, err := kubeLogin()
	if err != nil {
		glog.Errorf("ERROR: Failed to login using Kubernetes auth: %v", err)
		return
	}

	s, err := fetchSecrets(*secrets, client)
	if err != nil {
		glog.Errorf("ERROR: failed to fetch secrets: %v", err)
		return
	}

	if err := writeSecrets(s, *out); err != nil {
		glog.Errorf("ERROR: failed to write %d secrets to %s: %v", len(s), *out, err)
	}
}
