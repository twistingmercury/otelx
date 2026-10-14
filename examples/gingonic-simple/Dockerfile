FROM golang:1.27.2-alpine AS builder

WORKDIR /build

COPY src/go.mod src/go.sum ./src/
COPY src/*.go ./src/
COPY src/internal/ ./src/internal/
COPY src/middleware/ ./src/middleware/

COPY examples/gingonic-simple/ ./examples/gingonic-simple/

WORKDIR /build/examples/gingonic-simple
RUN go mod download && go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o /out/gingonic-simple ./cmd/main/main.go

FROM scratch AS final

WORKDIR /app

COPY --from=builder /out/gingonic-simple /app/gingonic-simple

EXPOSE 8080 9090

ENTRYPOINT ["/app/gingonic-simple"]
