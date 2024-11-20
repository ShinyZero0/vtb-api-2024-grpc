#!/usr/bin/env sh

set -exuo pipefail
# Wait for CA
# sleep 20
init() {

# if [ ! -f "$ROOT_CA_FINGERPRINT_PATH" ]; then
# 	inotifywait -e close_write --include "$(basename $ROOT_CA_FINGERPRINT_PATH)" "$(dirname $ROOT_CA_FINGERPRINT_PATH)"
# fi

# step ca bootstrap --install --fingerprint "$(cat $ROOT_CA_FINGERPRINT_PATH)"
# Download the root certificate
# export STEP_FINGERPRINT=$(cat $FINGERPRINT_PATH)
# step ca root $CLIENT_CERTS_DIR/root_ca.crt

if [ $(step ca provisioner list | jq '[.[]|select(.name=="mock")]|length') -gt 0 ]; then
	return
fi

step certificate install $(step path)/certs/root_ca.crt
until curl $OIDC_ENDPOINT; do
	sleep 1
done

# Get token
# dockerip=$(ping -c 1 -q $COMMON_NAME | head -1 | awk '{print $1}'| tail -c +2 | rev | tail -c +3 | rev)
# STEP_TOKEN=$(step ca token $COMMON_NAME --san 127.0.0.1 --san $COMMON_NAME)
step ca provisioner add mock --type OIDC \
	--client-id 12345 \
	--client-secret hackme \
	--configuration-endpoint https://mock_oidc:9000/oidc/.well-known/openid-configuration
# Download the root certificate
# step ca certificate --token $STEP_TOKEN $COMMON_NAME CLIENT_CERTS_DIR/site.crt CLIENT_CERTS_DIR/site.key
}
until curl -k $STEP_CA_URL/healthcheck; do
	sleep 1
done

init
exec "$@"
