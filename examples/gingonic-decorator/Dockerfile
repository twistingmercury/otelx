FROM golang:1.24-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
COPY *.go ./
COPY internal/ ./internal/
COPY middleware/ ./middleware/

COPY examples/gingonic-decorator/ ./examples/gingonic-decorator/

WORKDIR /build/examples/gingonic-decorator
RUN go mod download && go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o /out/gingonic-decorator ./cmd/main/main.go

FROM scratch AS final

WORKDIR /app

COPY --from=builder /out/gingonic-decorator /app/gingonic-decorator

EXPOSE 8080 9090

ENTRYPOINT ["/app/gingonic-decorator"]
