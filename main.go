package main

import (
	"flag"
	"strconv"

	"github.com/golang/glog"
)

var url = flag.String("url", "http://127.0.0.1:8200", "the Vault server URL")
var kubeAuthPath = flag.String("kubeAuthPath", "kubernetes", "the Vault path for Kubernetes auth (e.g. kubernetes/prod)")
var kubeAuthRole = flag.String("kubeAuthRole", "", "the role to use for Vault Kubernetes authentication")
var jwt = flag.String("jwt", "", "the token to use for Vault authentication")
var secrets = flag.String("secrets", "", "a comma separated list of paths, keys, and variable names e.g (/secret/s1#k1#name, /secret/s1#k2#name, /secret/s2#k5#name")
var tokenPath = flag.String("tokenPath", "/var/run/secrets/kubernetes.io/serviceaccount/token", "location of token - used if a token is not provided.")
var out = flag.String("out", "/var/run/secrets/vault", "location to store the secrets fetched from Vault")
var cert = flag.String("cert", "", "public key to use for HTTPS connections")
var insecure = flag.Bool("insecure", false, "allow insecure HTTPS connections")

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

	if *url == "" || *secrets == "" || *kubeAuthRole == "" {
		glog.Infof("Usage: ")
		flag.Usage()
		return
	}

	if *jwt == "" {
		s, err := lookupJwt()
		if err != nil {
			glog.Errorf("ERROR: failed to retrieve JWT: %v", err)
		}

		*jwt = s
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

	glog.Infof("Done!")
}
