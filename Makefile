APP=seldep


.PHONY: build
build: clean
	DOCKER_BUILDKIT=1 docker build --target bin --output bin/ .
	#docker build --target bin --output bin/ .
	#docker build -t seldon .
	##-v $(pwd)/bin:/workdir/bin
	#docker run  seldon make compile --target bin --output bin/ .


.PHONY: compile
compile:
	ls -al
	CGO_ENABLED=0 GOOS=darwin go build -o bin/app
	cd bin
	pwd
	ls -al

.PHONY: clean
clean:
	rm -rf bin/