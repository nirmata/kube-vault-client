package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/glog"

	"github.com/hashicorp/vault/api"
)

type secret struct {
	name  string
	path  string
	key   string
	value string
}

func (s *secret) String() string {
	displayVal := ""
	if s.value != "" {
		displayVal = "[...]"
	}

	return fmt.Sprintf("(%s): %s %s=%s", s.name, s.path, s.key, displayVal)
}

func fetchSecrets(paths string, client *api.Client) ([]*secret, error) {
	secrets, err := parsePaths(paths)
	if err != nil {
		return nil, err
	}

	results := make([]*secret, 0)
	for _, s := range secrets {
		glog.V(6).Infof("Fetching secret path=%s key=%s", s.path, s.key)
		vals, err := fetchSecret(s, client)
		if err != nil {
			return nil, err
		}

		results = append(results, vals...)
	}

	return results, nil
}

func parsePaths(paths string) ([]*secret, error) {
	results := make([]*secret, 0)
	p := strings.Split(paths, ",")
	for _, e := range p {
		e = strings.TrimSpace(e)
		toks := strings.Split(e, "#")
		num := len(toks)
		if num == 0 || num > 3 {
			return nil, fmt.Errorf("Invalid entry %s", e)
		}

		s := &secret{}
		switch num {
		case 1:
			s.path = strings.TrimSpace(toks[0])

		case 2:
			s.path = strings.TrimSpace(toks[0])
			s.key = strings.TrimSpace(toks[1])
			s.name = strings.TrimSpace(toks[1])

		case 3:
			s.path = strings.TrimSpace(toks[0])
			s.key = strings.TrimSpace(toks[1])
			s.name = strings.TrimSpace(toks[2])
		}

		results = append(results, s)
	}

	glog.V(6).Infof("Parsed paths %+v", results)
	return results, nil
}

func fetchSecret(s *secret, client *api.Client) ([]*secret, error) {
	resp, err := client.Logical().Read(s.path)
	if err != nil {
		glog.V(6).Infof("Failed to fetch secret %s from %s", s.String(), client.Address())
		return nil, err
	}

	if resp == nil {
		glog.V(3).Infof("No entry found at path %s", s.path)
		return nil, fmt.Errorf("Secret (%s) not found", s.path)
	}

	// Vault v1 KV returns secrets in "data"
	// Vault v2 KV returns secrets in "data/data"
	data := resp.Data
	if resp.Data["data"] != nil {
		data = resp.Data["data"].(map[string]interface{})
	}

	glog.V(6).Infof("Found %d entries at path %s", len(data), s.path)

	if s.key != "" {
		val, ok := data[s.key]
		if !ok {
			glog.V(3).Infof("No entry found at path %s and key %s", s.path, s.key)
			return nil, fmt.Errorf("Secret (%s#%s) not found", s.path, s.key)
		}

		s.value = val.(string)
		glog.Infof("Got secret: %s", s.String())
		results := make([]*secret, 1, 1)
		results[0] = s
		return results, nil
	}

	results := make([]*secret, 0)
	for k, v := range data {
		strVal := v.(string)
		secretEntry := &secret{path: s.path, key: k, name: k, value: strVal}
		glog.Infof("Got secret: %s", secretEntry.String())
		results = append(results, secretEntry)
	}

	return results, nil
}

func writeSecrets(secrets []*secret, location string) error {
	dir := filepath.Dir(location)
	os.MkdirAll(dir, 0755)

	f, err := os.Create(location)
	if err != nil {
		return err
	}

	w := bufio.NewWriter(f)
	for _, s := range secrets {
		w.WriteString(fmt.Sprintf("%s=%s\n", s.name, s.value))
	}

	w.Flush()
	glog.Infof("Wrote %d secrets to %s", len(secrets), location)
	return nil
}
