# регистрация OIDC провайдера в smallstep

ID и секрет клиента настроены в mock сервере, нужно вставить свои
```sh
step ca provisioner add mock --type OIDC \
	--client-id 12345 \
	--client-secret hackme \
	--configuration-endpoint https://mock_oidc:9000/oidc/.well-known/openid-configuration
```

# получение сертификатов по SSO

используем созданный ранее провайдер
```sh
step ca certificate --provisioner mock someone mf.crt mf.key
```

следующая команда симулирует аутентификацию пользователя через OIDC (используется mock сервер OIDC)

```sh
docker compose -f yea.yml exec smallstep curl -L 'https://mock_oidc:9000/oidc/authorize?client_id=12345&code_challenge=MKsompFPHOHnzCr27WBNEPpP-zeFC40QBNB97Q3zq8s&code_challenge_method=S256&nonce=9ce6f8ec734a79a3ca082010b16796abe991aad9b5bef7e643326b097f9fe7ff&redirect_uri=http%3A%2F%2F127.0.0.1%3A39393&response_type=code&scope=openid+email&state=3NDTaWndYrWGRNpHvRMeiZGx7NrLnYi4'
```


# идентификация пользователя по сертификату

[код](./cmd/server/main.go#L162)

