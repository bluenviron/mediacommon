test-pkg:
	go test -v -race -coverprofile=coverage-pkg.txt ./pkg/...

test-nodocker: test-pkg

define DOCKERFILE_TEST
FROM $(BASE_IMAGE)
RUN apk add --no-cache make git gcc musl-dev
WORKDIR /s
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
endef
export DOCKERFILE_TEST

test:
	echo "$$DOCKERFILE_TEST" | docker build -q . -f - -t temp
	docker run --rm -it \
	--name temp \
	temp \
	make test-nodocker
