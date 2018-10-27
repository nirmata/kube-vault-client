package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
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
var terminationMessagePath = flag.String("terminationMessagePath", "/dev/termination-log", "(optional) termination message path")

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
		msg := fmt.Sprintf("ERROR: Invalid usage")
		terminate(35, msg)
	}

	if *jwt == "" {
		s, err := lookupJwt()
		if err != nil {
			msg := fmt.Sprintf("ERROR: failed to retrieve JWT: %v", err)
			terminate(36, msg)
		}

		*jwt = s
	}

	client, err := kubeLogin()
	if err != nil {
		msg := fmt.Sprintf("ERROR: Failed to login using Kubernetes auth: %v", err)
		terminate(37, msg)
	}

	s, err := fetchSecrets(*secrets, client)
	if err != nil {
		msg := fmt.Sprintf("ERROR: failed to fetch secrets: %v", err)
		terminate(38, msg)
	}

	if err := writeSecrets(s, *out); err != nil {
		msg := fmt.Sprintf("ERROR: failed to write %d secrets to %s: %v", len(s), *out, err)
		terminate(39, msg)
	}

	terminate(0, fmt.Sprintf("Wrote Vault secrets to %s", *out))
}

func terminate(code int, message string) {
	if code != 0 {
		glog.Errorf("Exit Code %d: %s", code, message)
	} else {
		glog.Infof("%s", message)
	}

	ioutil.WriteFile(*terminationMessagePath, []byte(message), 0644)
	os.Exit(code)
}
