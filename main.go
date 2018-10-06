package main

import (
	"flag"
	"strconv"

	"github.com/golang/glog"
	"github.com/hashicorp/vault/api"
)

var url = flag.String("url", "http://127.0.0.1:8200", "the Vault server URL")
var token = flag.String("jwt", "", "the token to use for Vault authentication")
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

	if *url == "" || *secrets == "" {
		glog.Infof("Usage: ")
		flag.Usage()
		return
	}

	if *token == "" {
		lookupJwt()
		if *token == "" {
			glog.Errorf("ERROR: failed to retrieve Vault access token")
			return
		}
	}

	glog.Infof("Connecting to Vault at %s", *url)
	client, err := api.NewClient(&api.Config{
		Address: *url,
	})

	if err != nil {
		glog.Errorf("ERROR: failed to connect to Vault at %s: %v", *url, err)
		return
	}

	client.SetToken(*token)

	s, err := fetchSecrets(*secrets, client)
	if err != nil {
		glog.Errorf("ERROR: failed to fetch secrets: %v", err)
		return
	}

	if err := writeSecrets(s, *out); err != nil {
		glog.Errorf("ERROR: failed to write %d secrets to %s: %v", len(s), *out, err)
	}
}
