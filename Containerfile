FROM docker.io/library/golang:1.21 AS builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . ./
RUN go build -v -o /os-proxy .

FROM registry.access.redhat.com/ubi8/ubi-minimal

COPY --from=builder /os-proxy /os-proxy

ENTRYPOINT ["/os-proxy"]
