#!/usr/bin/env sh

set -exo pipefail
# Wait for CA
sleep 5

# Clean old certificates
rm -f /var/local/step/root_ca.crt
rm -f /var/local/step/site.crt /var/local/step/site.key

# Download the root certificate
# export STEP_FINGERPRINT=$(cat $FINGERPRINT_PATH)
step ca root /var/local/step/root_ca.crt

# Get token
STEP_TOKEN=$(step ca token $COMMON_NAME)
# Download the root certificate
step ca certificate --token $STEP_TOKEN $COMMON_NAME /var/local/step/site.crt /var/local/step/site.key

exec "$@"