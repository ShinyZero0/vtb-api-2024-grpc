BINS = bin/server bin/client bin/authserver
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
root.pem: cfssl.json
	# cfssl selfsign -config $< --profile rootca "Dev Testing CA" csr.json | cfssljson -bare root
	cfssl gencert -initca csr.json | cfssljson -bare root
%.pem: csr.json root.pem
	cfssl genkey $< | cfssljson -bare $(@:%.pem=%)
	cfssl sign -ca root.pem -ca-key root-key.pem -config cfssl.json -profile $(@:%.pem=%) $(@:%.pem=%).csr | cfssljson -bare $(@:%.pem=%)

# cfssl sign -ca root.pem -ca-key root-key.pem -config cfssl.json -profile client client.csr | cfssljson -bare client

CERTS = root.pem client.pem server.pem
certs: $(CERTS)
clean:
	git clean -xf
	# rm $(CERTS)
