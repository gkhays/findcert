.PHONY: build install clean test integration dep release docker
VERSION=`egrep -o '[0-9]+\.[0-9a-z.\-]+' version.go`
GIT_SHA=`git rev-parse --short HEAD || echo`

build:
	@echo "Building findcert..."
	@mkdir -p bin
	@go build -ldflags "-X main.GitSHA=${GIT_SHA}" -o bin/findcert .

install:
	@echo "Installing findcert..."
	@install -c bin/findcert /usr/local/bin/findcert

clean:
	@rm -f bin/*

dep:
	@dep ensure

release:
	@docker build -q -t jams_builder -f docker/Dockerfile.build.alpine .
	@for platform in darwin linux windows; do \
		if [ $$platform == windows ]; then extension=.exe; fi; \
		docker run -it --rm -v ${PWD}:/app -e "GOOS=$$platform" -e "GOARCH=amd64" -e "CGO_ENABLED=0" jams_builder go build -ldflags="-s -w -X main.GitSHA=${GIT_SHA}" -o bin/findcert-${VERSION}-$$platform-amd64$$extension; \
	done
	@docker run -it --rm -v ${PWD}:/app -e "GOOS=linux" -e "GOARCH=arm64" -e "CGO_ENABLED=0" jams_builder go build -ldflags="-s -w -X main.GitSHA=${GIT_SHA}" -o bin/findcert-${VERSION}-linux-arm64;
	@upx bin/findcert-${VERSION}-*

docker:
	@docker build --build-arg VERSION=${VERSION} -f docker/Dockerfile -t findcert:${VERSION} .