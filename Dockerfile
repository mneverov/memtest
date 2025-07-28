FROM golang:1.24 AS builder
WORKDIR /build/src
COPY go.mod go.mod
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

COPY main.go main.go

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -a -o memtest main.go

FROM alpine
WORKDIR /
COPY --from=builder /build/src/memtest /usr/bin/memtest

ENTRYPOINT ["/usr/bin/memtest"]
