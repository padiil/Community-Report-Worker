FROM golang:1.24-alpine AS builder
RUN apk add --no-cache build-base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -o ./worker ./cmd/worker

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/worker ./worker
RUN chown -R nobody:nobody /app
USER nobody
CMD ["./worker"]