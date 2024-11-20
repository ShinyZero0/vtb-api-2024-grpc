#!/usr/bin/env sh

set -exu

while sleep 1; do
	ready=1
	for file in $CAFILE $CERTFILE $KEYFILE; do
		[ ! -f $file ] && ready=0
	done
	[ $ready -eq 1 ] && break
done

exec "$@"
