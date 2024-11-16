BINS = bin/server bin/client bin/authserver bin/oidc-server-mock
all: $(BINS)
bin/client: generated-proto
bin/server: generated-proto server-models
generated-proto: main.proto
	mkdir -p $@
	protoc --go_out=$@ --go_opt=paths=source_relative \
		--go-grpc_out=$@ --go-grpc_opt=paths=source_relative \
		$<
	touch $@

bin/%: cmd/%/main.go **/*.go go.mod go.sum
	mkdir -p $(@D)
	go build -o $@ ./$(<D)

server-models: server-sqlboiler.toml server.db
	sqlboiler --wipe sqlite3 -c $< -o $@

%.db: dbschema.%.sqlite.sql
	sqlite3 $@ -init $< .exit

root.key:
	openssl genrsa -out $@ 2048
root.pem: root.key
	openssl req -config root.conf -new -x509 -key $< -out $@
%.conf: sanstr.txt
	cat /etc/ssl/openssl.cnf $<
%.csr %.key: %.conf
	openssl req -new -nodes -newkey rsa:4096 -keyout $*.key -out $@ -batch -subj "/C=DE/ST=Hamburg/L=Hamburg/O=Patrick CA/OU=router/CN=$*.box" -reqexts SAN -config $<

%.pem: root.key root.pem %.csr %.key
	openssl x509 -req -in $*.req -CA root.pem -CAkey root.key -CAcreateserial -out $@ -days 3650 -sha256 -extfile <(printf "subjectAltName=DNS:localhost,IP:127.0.0.1")

%-key.pem: %.key
	cp $< $@
# cfssl sign -ca root.pem -ca-key root-key.pem -config cfssl.json -profile client client.csr | cfssljson -bare client

CERTS = root.pem client.pem server.pem client-key.pem server-key.pem
.PRECIOUS: %.key
certs: $(CERTS)
clean:
	git clean -xf
	# rm $(CERTS)
