#!/usr/bin/env sh

CAFILE=/home/step/certs/root_ca.crt
until [ -f "$CAFILE" ]; do
	sleep 1
	# inotifywait -e close_write --include "$(basename $CAFILE)" "$(dirname $CAFILE)"
done
step certificate fingerprint $CAFILE > $ROOT_CA_FINGERPRINT_PATH
