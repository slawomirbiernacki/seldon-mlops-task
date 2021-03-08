PLATFORM=local

.PHONY: build
build: clean
	make compile-in-docker

.PHONY: build-dev
build-dev: clean
	go test ./...
	go build -o bin/app

.PHONY: compile-in-docker
compile-in-docker:
	DOCKER_BUILDKIT=1 docker build --target bin --output bin/ --platform ${PLATFORM} .

.PHONY: clean
clean:
	rm -rf bin/