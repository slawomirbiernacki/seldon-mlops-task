APP=seldep

PLATFORM=local

.PHONY: build
build: clean
	DOCKER_BUILDKIT=1 docker build --target bin --output bin/ --platform ${PLATFORM} .

.PHONY: compile
compile:

	GOOS=${TARGETOS} GOARCH=${TARGETARCH}  go build -o bin/app-${TARGETOS}-${TARGETARCH}

.PHONY: clean
clean:
	rm -rf bin/