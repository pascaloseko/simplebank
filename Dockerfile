FROM golang:1.18.10-bullseye AS builder

ENV TZ=UTC \
    PATH=/root/go/bin:$PATH

WORKDIR /app

RUN go env -w GOPATH=/root/go GOBIN=/root/go/bin

ADD go.mod go.sum /app/
RUN go mod download

# Remove deprecated certificate
RUN rm /usr/share/ca-certificates/mozilla/DST_Root_CA_X3.crt \
  && sed -i '/DST_Root_CA_X3/d' /etc/ca-certificates.conf \
  && update-ca-certificates

###################################################
# install golangci for linting
FROM golangci/golangci-lint:v1.45.1 as golangci-lint

###################################################
# install sqlc for code generation
FROM kjconroy/sqlc:1.13.0 as sqlc

###################################################
# install golang-migrate for migrations cli
FROM migrate/migrate:v4.15.1 as golang-migrate

###################################################
FROM builder AS local

ENV TZ=UTC \
    PATH=/root/go/bin:$PATH

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    # edit pip installed libraries
    vim \
    # ps to investigate / kill hung processes
    procps

COPY --from=golangci-lint /usr/bin/golangci-lint /go/bin/golangci-lint

COPY --from=golang-migrate /usr/local/bin/migrate /go/bin/migrate

COPY --from=sqlc /workspace/sqlc /go/bin/sqlc

CMD ["go", "run", "."]

###################################################
FROM builder AS cloudbuilder

ENV TZ=UTC \
    PATH=/root/go/bin:$PATH

WORKDIR /app

COPY . /app

RUN go build -o /usr/local/bin

###################################################
FROM python:3.7.7-slim-buster AS cloud

WORKDIR /app

COPY --from=cloudbuilder /etc/ssl/certs /etc/ssl/certs

COPY --from=cloudbuilder /usr/local/bin/simple-bank /usr/local/bin/simple-bank

CMD ["/usr/local/bin/simple-bank"]

LABEL Description="Simple Bank Backend image" Vendor="Pascal Oseko"
