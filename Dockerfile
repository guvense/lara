FROM golang:1.17.5 AS build

LABEL MAINTAINER = 'Guven Seckin'

ARG VERSION=dev
WORKDIR /go/src/github.com/guvense/lara
COPY . .
RUN go mod tidy \
    && LDFLAGS=$(echo "-s -w -X main.version="${VERSION}) \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/lara -ldflags "$LDFLAGS" cmd/lara/main.go

# Building image with the binary
FROM scratch
COPY --from=build /go/bin/lara /go/bin/lara
WORKDIR /go/src/github.com/guvense/lara
ENTRYPOINT ["/go/bin/lara"]