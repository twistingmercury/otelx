FROM ghcr.io/twistingmercury/golang-tooling:latest AS verify

RUN apk add --no-cache gcc musl-dev

WORKDIR /workspace

COPY go.mod go.sum go.work ./
COPY examples/gingonic-simple/go.mod examples/gingonic-simple/go.sum ./examples/gingonic-simple/
COPY examples/gingonic-decorator/go.mod examples/gingonic-decorator/go.sum ./examples/gingonic-decorator/

RUN go mod download

COPY . .

RUN goimports -w .
RUN golangci-lint run
RUN govulncheck ./...
RUN gosec -quiet -exclude-dir=tests ./...

RUN go vet ./...
RUN CGO_ENABLED=1 go test -race ./...
RUN go build ./...
