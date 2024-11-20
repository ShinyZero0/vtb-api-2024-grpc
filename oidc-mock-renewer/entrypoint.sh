#!/usr/bin/env sh

set -exuo pipefail
# Wait for CA
# sleep 5

# init() {

# 	until [ -f $ROOT_CA_FINGERPRINT_PATH ]; do
# 		sleep 1
# 	done
# 	step ca bootstrap --install -f --fingerprint $(cat $ROOT_CA_FINGERPRINT_PATH)
# 	tree $(step path)
# }
# if [ ! -f "$(step path)/config/defaults.json" ]; then
# 	init
# fi


# Clean old certificates

until curl -k $STEP_CA_URL/healthcheck; do
	sleep 1
done
# Download the root certificate
# export STEP_FINGERPRINT=$(cat $FINGERPRINT_PATH)
rm -f $CLIENT_CERTS_DIR/root_ca.crt
rm -f $CLIENT_CERTS_DIR/site.crt $CLIENT_CERTS_DIR/site.key
step ca root $CLIENT_CERTS_DIR/root_ca.crt

# Get token
# dockerip=$(ping -c 1 -q $COMMON_NAME | head -1 | awk '{print $1}'| tail -c +2 | rev | tail -c +3 | rev)
STEP_TOKEN=$(step ca token $COMMON_NAME --san 127.0.0.1 --san $COMMON_NAME)
# Download the root certificate
step ca certificate --token $STEP_TOKEN $COMMON_NAME $CLIENT_CERTS_DIR/site.crt $CLIENT_CERTS_DIR/site.key

exec "$@"
