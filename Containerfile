# syntax=docker/dockerfile:1

FROM docker.io/library/golang:1.22 AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY ./pkg ./pkg
COPY ./cmd ./cmd

RUN CGO_ENABLED=0 GOOS=linux go build -o /openstack-proxy ./cmd/openstack-proxy

FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

ENV OS_CLOUD=openstack OS_CLIENT_CONFIG_FILE=/config/clouds.yaml

EXPOSE 13000

COPY --from=build /openstack-proxy /

ENTRYPOINT ["/openstack-proxy"]
