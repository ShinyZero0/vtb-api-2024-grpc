#!/bin/bash
set	-exo pipefail

# stealed from:
# https://github.com/smallstep/certificates/blob/master/docker/entrypoint.sh

export STEPPATH=$(step path)

# List of env vars required	for	step ca	init
declare	-ra	REQUIRED_INIT_VARS=(DOCKER_STEPCA_INIT_NAME	DOCKER_STEPCA_INIT_DNS_NAMES)

# Ensure all env vars required to run step ca init are set.
function init_if_possible () {
	local missing_vars=0
	for	var	in "${REQUIRED_INIT_VARS[@]}"; do
		if [ -z	"${!var}" ]; then
			missing_vars=1
		fi
	done
	if [ ${missing_vars} = 1 ];	then
		>&2	echo "there	is no ca.json config file; please run step ca init,	or provide config parameters via DOCKER_STEPCA_INIT_ vars"
	else
		step_ca_init "${@}"
	fi
}

function generate_password () {
	set	+o pipefail
	< /dev/urandom tr -dc A-Za-z0-9	| head -c40
	echo
	set	-o pipefail
}

# Initialize a CA if not already initialized
function step_ca_init () {
	DOCKER_STEPCA_INIT_PROVISIONER_NAME="${DOCKER_STEPCA_INIT_PROVISIONER_NAME:-admin}"
	DOCKER_STEPCA_INIT_ADMIN_SUBJECT="${DOCKER_STEPCA_INIT_ADMIN_SUBJECT:-step}"
	DOCKER_STEPCA_INIT_ADDRESS="${DOCKER_STEPCA_INIT_ADDRESS:-:9000}"

	local -a setup_args=(
		--name "${DOCKER_STEPCA_INIT_NAME}"
		--dns "${DOCKER_STEPCA_INIT_DNS_NAMES}"
		--provisioner "${DOCKER_STEPCA_INIT_PROVISIONER_NAME}"
		--password-file	"${STEPPATH}/password"
		--provisioner-password-file	"${STEPPATH}/provisioner_password"
		--address "${DOCKER_STEPCA_INIT_ADDRESS}"
	)
	if [ -n	"${DOCKER_STEPCA_INIT_PASSWORD_FILE}" ]; then
		< "${DOCKER_STEPCA_INIT_PASSWORD_FILE}"	tee "${STEPPATH}/password" "${STEPPATH}/provisioner_password"
	elif [ -n "${DOCKER_STEPCA_INIT_PASSWORD}" ]; then
		echo "${DOCKER_STEPCA_INIT_PASSWORD}" | tee "${STEPPATH}/password" "${STEPPATH}/provisioner_password"
	else
		generate_password >	"${STEPPATH}/password"
		generate_password >	"${STEPPATH}/provisioner_password"
	fi
	if [ "${DOCKER_STEPCA_INIT_SSH}" ==	"true" ]; then
		setup_args=("${setup_args[@]}" --ssh)
	fi
	if [ "${DOCKER_STEPCA_INIT_ACME}" == "true"	]; then
		setup_args=("${setup_args[@]}" --acme)
	fi
	if [ "${DOCKER_STEPCA_INIT_REMOTE_MANAGEMENT}" == "true" ];	then
		setup_args=(
		"${setup_args[@]}"
		--remote-management
		--admin-subject "${DOCKER_STEPCA_INIT_ADMIN_SUBJECT}")
	fi

	step ca	init "${setup_args[@]}"
	# step ca bootstrap --install --fingerprint $fingerprint

	# if [ -n "$DOCKER_STEPCA_INIT_OIDC_ENDPOINT" ]; then
	# 	( sleep 15
	# 	step ca provisioner add mock --type OIDC \
	# 		--client-id 12345 \
	# 		--client-secret hackme \
	# 		--configuration-endpoint $DOCKER_STEPCA_INIT_OIDC_ENDPOINT) &
	# fi
	echo ""
	if [ "${DOCKER_STEPCA_INIT_REMOTE_MANAGEMENT}" == "true" ];	then
		echo "👉 Your CA administrative	username is: ${DOCKER_STEPCA_INIT_ADMIN_SUBJECT}"
	fi
	echo "👉 Your CA administrative	password is: $(< $STEPPATH/provisioner_password	)"
	echo "🤫 This will only	be displayed once."
	shred -u $STEPPATH/provisioner_password
	mv $STEPPATH/password $PWDPATH
}

if [ -f	/usr/sbin/pcscd	]; then
	/usr/sbin/pcscd
fi

if [ ! -f "${STEPPATH}/config/ca.json" ]; then
	init_if_possible
else
	tree ${STEPPATH}
	until curl $OIDC_ENDPOINT; do
		sleep 1
	done
fi
# if [ $(step ca provisioner list | jq '[.[]|select(.name=="mock")]|length') -gt 0 ]; then
# fi

exec "${@}"
