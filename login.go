package main

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"path"
	"strings"

	"github.com/golang/glog"
	"github.com/hashicorp/vault/api"
)

func kubeLogin() (*api.Client, error) {
	glog.Infof("Connecting to Vault at %s", *url)

	httpClient := buildHTTPClient(*url)
	config := &api.Config{
		Address:    *url,
		HttpClient: httpClient,
	}

	client, err := api.NewClient(config)
	if err != nil {
		glog.Errorf("ERROR: failed to connect to Vault at %s: %v", *url, err)
		return nil, err
	}

	body := map[string]interface{}{
		"role": *kubeAuthRole,
		"jwt":  *jwt,
	}

	loginPath := "/v1/auth/" + *kubeAuthPath + "/login"
	loginPath = path.Clean(loginPath)
	glog.Infof("Vault login using path %s role %s jwt [%d bytes]", loginPath, *kubeAuthRole, len(*jwt))

	req := client.NewRequest("POST", loginPath)
	req.SetJSONBody(body)

	resp, err := client.RawRequest(req)
	if err != nil {
		glog.Errorf("ERROR: failed to login with Vault: %v", err)
		return nil, err
	}

	if respErr := resp.Error(); respErr != nil {
		glog.Errorf("ERROR: api error: %v", respErr)
		return nil, err
	}

	var result api.Secret
	if err := resp.DecodeJSON(&result); err != nil {
		glog.Errorf("ERROR: failed to decode JSON response: %v", err)
		return nil, err
	}

	glog.Infof("Login results %+v", result)

	client.SetToken(result.Auth.ClientToken)
	return client, nil
}

func buildHTTPClient(url string) *http.Client {

	if strings.HasPrefix(url, "http://") {
		return http.DefaultClient
	}

	sslCerts := CACERTS
	if *cert != "" {
		sslCerts = []byte(*cert)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(sslCerts)
	tlsConfig := &tls.Config{RootCAs: caCertPool, InsecureSkipVerify: *insecure}
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	httpClient := &http.Client{Transport: transport}
	return httpClient
}
