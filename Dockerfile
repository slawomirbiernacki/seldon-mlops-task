FROM --platform=${BUILDPLATFORM} golang:latest AS build

WORKDIR /workdir

COPY go.mod .
COPY go.sum .
RUN go mod download -x

COPY . .

ENV GO111MODULE=on
ENV CGO_ENABLED=0

ARG TARGETOS
ARG TARGETARCH

# This takes a while every time - because of CGO_ENABLED=0 it recompiles all dependencies. Couldn't find a way to cache that before building my sources when using modules
#RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH}  go build -o bin/seldon-mlops-task-${TARGETOS}-${TARGETARCH}
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH}  go build -a -tags netgo -installsuffix netgo -ldflags "-linkmode external -extldflags -static" -o bin/seldon-mlops-task-${TARGETOS}-${TARGETARCH} *.go

FROM scratch AS bin
COPY --from=build /workdir/bin/ /