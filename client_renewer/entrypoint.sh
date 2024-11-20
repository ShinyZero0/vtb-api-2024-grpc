#!/usr/bin/env bash

set -exuo pipefail
# Wait for CA
# sleep 5

# Clean old certificates
rm -f $CLIENT_CERTS_DIR/root_ca.crt
rm -f $CLIENT_CERTS_DIR/site.crt $CLIENT_CERTS_DIR/site.key

until [ -f $ROOT_CA_FINGERPRINT_PATH ]; do
	sleep 1
done

until curl -k $STEP_CA_URL/healthcheck; do
	sleep 1
done
step ca bootstrap -f --install --fingerprint $(cat $ROOT_CA_FINGERPRINT_PATH)
# Download the root certificate
# export STEP_FINGERPRINT=$(cat $FINGERPRINT_PATH)
step ca root $CLIENT_CERTS_DIR/root_ca.crt

until [ $(step ca provisioner list | jq '[.[]|select(.name=="mock")]|length') -gt 0 ]; do
	sleep 1
done


# Get token
# dockerip=$(ping -c 1 -q $COMMON_NAME | head -1 | awk '{print $1}'| tail -c +2 | rev | tail -c +3 | rev)
# STEP_TOKEN=$(step ca token $COMMON_NAME --san 127.0.0.1 --san $COMMON_NAME)
step ca certificate --provisioner mock someone $CLIENT_CERTS_DIR/site.crt $CLIENT_CERTS_DIR/site.key |& xargs -rn 1 curl -L ||:
# Download the root certificate
# step ca certificate --token $STEP_TOKEN $COMMON_NAME $CLIENT_CERTS_DIR/site.crt $CLIENT_CERTS_DIR/site.key

exec "$@"
