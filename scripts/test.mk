LBITS := $(shell getconf LONG_BIT)
ifeq ($(LBITS),64)
RACE=-race
endif

test-pkg:
	go test -v $(RACE) -coverprofile=coverage-pkg.txt ./pkg/...

test-nodocker: test-pkg

define DOCKERFILE_TEST
ARG ARCH
FROM --platform=$$ARCH $(BASE_IMAGE)
RUN apk add --no-cache make git gcc musl-dev
WORKDIR /s
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
endef
export DOCKERFILE_TEST

test:
	echo "$$DOCKERFILE_TEST" | docker build -q . -f - -t temp --build-arg ARCH=amd64
	docker run --rm -it \
	--name temp \
	temp \
	make test-nodocker

test-32:
	echo "$$DOCKERFILE_TEST" | docker build -q . -f - -t temp --build-arg ARCH=i386
	docker run --rm \
	--name temp \
	temp \
	make test-nodocker
