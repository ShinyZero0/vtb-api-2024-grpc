#!/usr/bin/env bash
# export SERVER_ADDR=:9090
export LISTEN_ADDR=localhost:9098
export DSN=server.db
# export JWT_SECRET=aaaaaaaaaah
./bin/authserver

