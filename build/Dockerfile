FROM ghcr.io/twistingmercury/golang-tooling:go1.27.2 AS verify

RUN apk add --no-cache gcc musl-dev

WORKDIR /workspace

COPY go.mod go.sum go.work ./
COPY middleware/gin/go.mod middleware/gin/go.sum ./middleware/gin/
COPY examples/gingonic-simple/go.mod examples/gingonic-simple/go.sum ./examples/gingonic-simple/
COPY examples/gingonic-decorator/go.mod examples/gingonic-decorator/go.sum ./examples/gingonic-decorator/

RUN go mod download

COPY . .

RUN goimports -w .

# The library and the gin middleware are separate modules; check both.
ARG MODULES=". ./middleware/gin"

RUN for m in $MODULES; do (cd "$m" && golangci-lint run) || exit 1; done
RUN for m in $MODULES; do (cd "$m" && govulncheck ./...) || exit 1; done
RUN for m in $MODULES; do (cd "$m" && gosec -quiet -exclude-dir=tests ./...) || exit 1; done

RUN for m in $MODULES; do (cd "$m" && go vet ./...) || exit 1; done
RUN for m in $MODULES; do (cd "$m" && CGO_ENABLED=1 go test -race ./...) || exit 1; done
RUN for m in $MODULES; do (cd "$m" && go build ./...) || exit 1; done
