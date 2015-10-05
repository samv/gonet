#!/usr/bin/env bash

export name=`basename $1`
go test -race -coverprofile ./${name}.cover -c -v $1
sudo setcap CAP_NET_RAW=epi ./${name}.test
sudo setcap CAP_NET_ADMIN=epi ./${name}.test
./${name}.test -test.v -test.coverprofile ./${name}.cover
#rm ${name}.test
go tool cover -html=${name}.cover -o ${name}.html
