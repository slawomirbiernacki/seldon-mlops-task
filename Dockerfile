FROM --platform=${BUILDPLATFORM} golang:1.14.3-alpine AS build

WORKDIR /workdir

COPY Makefile .
COPY main.go .
COPY go.mod .
COPY go.sum .
ADD seldonclient seldonclient

ENV GO111MODULE=on
ENV CGO_ENABLED=0

ARG TARGETOS
ARG TARGETARCH

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH}  go build -o bin/app-${TARGETOS}-${TARGETARCH}

FROM scratch AS bin
COPY --from=build /workdir/bin/ /