FROM golang:1.14 AS build

WORKDIR /workdir

COPY Makefile .
COPY main.go .
COPY go.mod .
COPY go.sum .
ADD seldonclient seldonclient

ENV GO111MODULE=on

RUN CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o bin/app

FROM scratch AS bin
COPY --from=build /workdir/bin/app /