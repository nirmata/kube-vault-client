package main

import (
	"path"

	"github.com/golang/glog"
	"github.com/hashicorp/vault/api"
)

func kubeLogin() (*api.Client, error) {
	glog.Infof("Connecting to Vault at %s", *url)
	config := &api.Config{
		Address: *url,
	}

	tls := &api.TLSConfig{Insecure: true}
	config.ConfigureTLS(tls)

	client, err := api.NewClient(config)
	if err != nil {
		glog.Errorf("ERROR: failed to connect to Vault at %s: %v", *url, err)
		return nil, err
	}

	body := map[string]interface{}{
		"role": *role,
		"jwt":  *jwt,
	}

	loginPath := "/v1/auth/" + *kubeAuthPath + "/login"
	loginPath = path.Clean(loginPath)
	glog.Infof("Vault login using path %s role %s jwt [%d bytes]", loginPath, *role, len(*jwt))

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

	auth := result.Auth
	glog.Infof("Got auth %+v", auth)

	client.SetToken(auth.ClientToken)
	return client, nil
}
