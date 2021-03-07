PLATFORM=local

.PHONY: build
build: clean
	DOCKER_BUILDKIT=1 docker build --target bin --output bin/ --platform ${PLATFORM} .


.PHONY: build-all
build-all: clean
	make build PLATFORM=linux/amd64
	make build PLATFORM=darwin/amd64

.PHONY: build-dev
build-dev: clean
	go build -o bin/app

.PHONY: clean
clean:
	rm -rf bin/