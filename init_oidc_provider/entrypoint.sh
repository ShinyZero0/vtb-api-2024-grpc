#!/usr/bin/env sh

set -exuo pipefail
# Wait for CA
# sleep 20
init() {
if [ $(step ca provisioner list | jq '[.[]|select(.name=="mock")]|length') -gt 0 ]# Clean old certificates; then
	return
fi


rm -f /var/local/step/root_ca.crt
rm -f /var/local/step/site.crt /var/local/step/site.key

if [ ! -f "$ROOT_CA_FINGERPRINT_PATH" ]; then
	inotifywait -e close_write --include "$(basename $ROOT_CA_FINGERPRINT_PATH)" "$(dirname $ROOT_CA_FINGERPRINT_PATH)"
fi

until curl -k $STEP_CA_URL/healthcheck; do
	sleep 1
done

step ca bootstrap --install --fingerprint "$(cat $ROOT_CA_FINGERPRINT_PATH)"
# Download the root certificate
# export STEP_FINGERPRINT=$(cat $FINGERPRINT_PATH)
# step ca root $CLIENT_CERTS_DIR/root_ca.crt

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
init
exec "$@"
