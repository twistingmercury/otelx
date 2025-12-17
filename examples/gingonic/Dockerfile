FROM golang:1.24-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
COPY *.go ./
COPY internal/ ./internal/
COPY middleware/ ./middleware/

COPY examples/gingonic/ ./examples/gingonic/

WORKDIR /build/examples/gingonic
RUN go mod download && go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o /out/gin-gonic-example ./cmd/main/main.go

FROM scratch AS final

WORKDIR /app

COPY --from=builder /out/gin-gonic-example /app/gin-gonic-example

EXPOSE 8080 9090

ENTRYPOINT ["/app/gin-gonic-example"]
