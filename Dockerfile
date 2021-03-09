FROM --platform=${BUILDPLATFORM} golang:1.14.3-alpine AS build

WORKDIR /workdir

COPY go.mod .
COPY go.sum .
RUN go mod download -x

COPY . .

ENV GO111MODULE=on
ENV CGO_ENABLED=0

ARG TARGETOS
ARG TARGETARCH

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH}  go build -o bin/seldon-mlops-task-${TARGETOS}-${TARGETARCH}

FROM scratch AS bin
COPY --from=build /workdir/bin/ /