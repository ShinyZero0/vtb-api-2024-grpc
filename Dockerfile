FROM golang:1.23.2

RUN apt-get update
RUN apt-get install -y protobuf-compiler
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest 
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
RUN go install github.com/volatiletech/sqlboiler/v4@latest
RUN go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-sqlite3@latest
RUN apt-get install -y sqlite3

WORKDIR /project
COPY go.mod go.sum ./
RUN go mod download
ADD . .
RUN make
