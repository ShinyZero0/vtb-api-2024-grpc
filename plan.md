# resource server

## verifying tokens

### introspection

- <https://github.com/bploetz/caddy-oauth2-token-introspection>
- <https://pkg.go.dev/github.com/shaj13/go-guardian/v2/auth/strategies/oauth2/introspection
- <https://datatracker.ietf.org/doc/html/rfc7662>

### handmade crap


# authn


## mtls

yes, we are our own CA

### identifying certs
`x5t#S256` from <https://datatracker.ietf.org/doc/html/rfc8705#section-3.1-1>

### automation?

need some service to issue certs (semi)automatically in a scalable manner

#### cfssl

#### smallstep

<https://blog.xentoo.info/2021/09/12/running-a-pki-using-smallstep-certificates-with-docker/>
- <https://github.com/smallstep/certificates>
- <https://blog.xentoo.info/2021/09/12/running-a-pki-using-smallstep-certificates-with-docker/>
- <https://smallstep.com/docs/step-ca/provisioners/#unattended-remote-provisioner-management>

## oidc for client ui

### mocks

- <https://github.com/oauth2-proxy/mockoidc>

## combine?

separate ports, one with mtls, another unencrypted (for encrypting with sidecar proxy) and token-authenticated

<https://datatracker.ietf.org/doc/html/rfc8705#name-mutual-tls-client-certifica>

# RBAC

## keto

# antiddos

## haproxy
<https://www.haproxy.com/blog/application-layer-ddos-attack-protection-with-haproxy>

### mtls
<https://webhostinggeeks.com/howto/how-to-configure-haproxy-with-ssl-pass-through/>

# antivirus

## yara

- <https://github.com/VirusTotal/gyp>
- <https://github.com/Neo23x0/signature-base>

# performance

## multithreaded what?
chat

## streaming
grpc

streams + cert expiry
https://github.com/grpc/grpc-go/issues/3021

# business logic

doin a chat

## spamming with bots

- <https://superuser.openinfra.dev/articles/run-load-balanced-service-docker-containers-openstack/>

### docker

- compose scaling/replicating:
<https://docs.docker.com/reference/compose-file/deploy/#replicas>
