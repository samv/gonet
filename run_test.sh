#!/usr/bin/env bash

export name=`basename $1`
go test -race -c -v $1
if [ ! -f /.dockerenv ]; then
	# not inside docker
	# see https://github.com/docker/docker/issues/5650
	sudo setcap CAP_NET_RAW=epi ./${name}.test
	sudo setcap CAP_NET_ADMIN=epi ./${name}.test
fi
./${name}.test -test.v
#rm ${name}.test
