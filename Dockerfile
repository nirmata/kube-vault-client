FROM scratch
LABEL Author "Jim Bugwadia <jim@nirmata.com>"

ADD vault-client /vault-client
ENTRYPOINT ["/vault-client"]