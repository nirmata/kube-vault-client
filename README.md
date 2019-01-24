# kube-vault-client

A minimal Vault client to manage secrets for Kubernetes pods.

***Note: This project is in alpha stage. We are actively working on improving the functionality and incorporating the user feedback. Please see the roadmap. You are welcome to tryout and provide feedback.***


## Overview

[Vault](https://www.vaultproject.io/) by HashiCorp is a tool for managing secrets i.e. sensitive data like passwords and certificates. Vault provides advanced features like dynamic secrets management and leases. 
Vault also integrates nicely with Kubernetes to allow using Kubernetes [Service Accounts](https://kubernetes.io/docs/reference/access-authn-authz/service-accounts-admin/) for authentictaion and authorization. 

Vault can be accessed via an API or a command line tool. However, its typically not desireable, nor 
practical, to update application code to access Vault. 

*kube-vault-client* is a Golang program, built as a static container, and designed to run in a Kubernetes [Pod](https://kubernetes.io/docs/concepts/workloads/pods/pod-overview/), to manage the process of fetching secrets from Vault. 

Here is what it does:
1. Uses the pod's service account JWT token and a supplied role to authenticate with Vault using the Kubernetes Authentication method.
2. Once authenticated, fetches the secrets specified in a supplied path. The secrets are written to a configured file location.

You can read more how Nirmata uses *kube-vault-client* in this [blog post](https://www.nirmata.com/2018/12/19/managing-kubernetes-secrets-with-hashicorp-vault-and-nirmata/).


## Approach

**kube-vault-client** focuses on retrieving secrets from Vault. It is designed to be minimal, which 
is always great for security, but complete. This allows operational concerns for Vault, to be cleanly separated from the concerns of accessing secrets.

The design of **kube-vault-client** was heavily inspired by [kube-vault-auth-init] (https://github.com/WealthWizardsEngineering/kube-vault-auth-init). However, **kube-vault-client** does not require a separate AppRole, provides flexible options to map secrets, supports namespaces, and provides detailed status and error reporting for use as an Kuberneres init container.

## Usage

### Running locally

For local testing, you can run the image from docker hub

````bash
docker run nirmata/kube-vault-client:2.3.0
````

Running the image with no parameters will display the usage:

````bash
Usage of /kube-vault-client:
  -alsologtostderr
        log to standard error as well as files
  -cert string
        public key to use for HTTPS connections
  -insecure
        allow insecure HTTPS connections
  -jwt string
        the token to use for Vault authentication
  -kubeAuthPath string
        the Vault path for Kubernetes auth (e.g. kubernetes/prod) (default "kubernetes")
  -kubeAuthRole string
        the role to use for Vault Kubernetes authentication
  -log_backtrace_at value
        when logging hits line file:N, emit a stack trace
  -log_dir string
        If non-empty, write log files in this directory
  -logtostderr
        log to standard error instead of files
  -namespace
        (optional) the namespace to use if you have Vault Enterprise (X-Vault-Namespace)
  -out string
        location to store the secrets fetched from Vault (default "/var/run/secrets/vault")
  -secrets string
        a comma separated list of paths, keys, and variable names e.g (/secret/s1#k1#name, /secret/s1#k2#name, /secret/s2#k5#name
  -stderrthreshold value
        logs at or above this threshold go to stderr
  -terminationMessagePath string
        (optional) termination message path (default "/dev/termination-log")
  -tokenPath string
        location of token - used if a token is not provided. (default "/var/run/secrets/kubernetes.io/serviceaccount/token")
  -url string
        the Vault server URL (default "http://127.0.0.1:8200")
  -v value
        log level for V logs
  -vmodule value
        comma-separated list of pattern=N settings for file-filtered logging
````

#### The secrets string

You can control which secrets are retrieved from Vault using a comma separated list of paths. To retrieve a single entry, you can provide the entry key. For example:

````
secret/certs, secret/mysql#password
````

If the key is not specified, all keys at the specified path are retrieved. If a key is specified, you can also provide an optional variable name. For example:

````
secret/mysql#password#MYSQL_PASSWORD
````

This variable name is used when the secret is stored in the location specified by the **out** parameter. 

### Running in a Kubernetes Cluster

To fetch secrets for a Kubernetes application, you can run *kube-vault-client* as an init container within a Kubernetes pod.

Here is an example for running 

````yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: ghost
  namespace: ghost-aws-demo-prod
spec:
  replicas: 1
  revisionHistoryLimit: 5
  selector:
    matchLabels:
      nirmata.io/component: ghost
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
    type: RollingUpdate
  template:
    metadata:
      creationTimestamp: null
      labels:
        nirmata.io/component: ghost
    spec:
      containers:
      - env:
        - name: VAULT_SECRETS_FILEPATH
          value: /var/run/vault
        image: ghost:0.11.9-alpine
        imagePullPolicy: Always
        name: ghost
        ports:
        - containerPort: 2368
          protocol: TCP
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /var/run
          name: vault-secrets
      dnsPolicy: ClusterFirst
      initContainers:
      - args:
        - -kubeAuthRole
        - ghost-prod-role
        - -out
        - /var/run/vault
        - -secrets
        - secret/ghost/prod, secret/ghost/prod#password#MY_PASSWORD, secret/ghost/prod#token#MY_TOKEN
        - -url
        - https://vault-devtest2.nirmata.io/
        - -kubeAuthPath
        - kubernetes/prod/aws-demo/
        image: docker.io/nirmata/kube-vault-client:latest
        name: vault-init-secrets
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /var/run
          name: vault-secrets
      serviceAccount: ghost-prod
      serviceAccountName: ghost-prod
      terminationGracePeriodSeconds: 30
      volumes:
      - hostPath:
          path: /var/run
          type: ""
        name: vault-secrets
````

## Exit Codes

To propogate and report errors to Kubernetes APIs and utilities *kube-vault-client* uses the following exit codes and writes a message to the path specified in the *terminationMessagePath* command.

| Code  | Meaning                      |
| :---: | :---                         |
| 0     | Success                      |
| 35    | Invalid usage                |
| 36    | Missing or invalid JWT token |
| 37    | Error during Vault login     |
| 38    | Error while reading secrets  |
| 39    | Error while writing secrets  |


## Building

To build the binary and docker image, clone this repository and run:

````bash
make 
````

## Roadmap

- support periodic renewal of secrets that have an associated lease



