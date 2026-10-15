FROM ghcr.io/twistingmercury/golang-tooling:go1.27.2 AS verify

RUN apk add --no-cache gcc musl-dev

WORKDIR /workspace

COPY src/go.mod src/go.sum src/go.work ./src/
COPY src/middleware/gin/go.mod src/middleware/gin/go.sum ./src/middleware/gin/
COPY examples/gingonic-simple/go.mod examples/gingonic-simple/go.sum ./examples/gingonic-simple/
COPY examples/gingonic-decorator/go.mod examples/gingonic-decorator/go.sum ./examples/gingonic-decorator/

WORKDIR /workspace/src

RUN go mod download

WORKDIR /workspace

COPY . .

WORKDIR /workspace/src

## Run linters, formatters, security scanners, etc
RUN goimports -w .
RUN golangci-lint run
RUN govulncheck ./...
RUN gosec ./...

WORKDIR /workspace

# The library and the gin middleware are separate modules; check both.
ARG MODULES="src src/middleware/gin"

RUN for m in $MODULES; do (cd "$m" && go vet ./...) || exit 1; done
RUN for m in $MODULES; do (cd "$m" && CGO_ENABLED=1 go test -race ./...) || exit 1; done
RUN for m in $MODULES; do (cd "$m" && go build ./...) || exit 1; done
