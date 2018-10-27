FROM scratch
LABEL Author "Jim Bugwadia <jim@nirmata.com>"

ADD kube-vault-client /kube-vault-client
ENTRYPOINT ["/kube-vault-client"]