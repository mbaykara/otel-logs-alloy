# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o dice-app ./cmd/dice-app

# Final stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/dice-app .

# Expose ports for HTTP server and OTLP
EXPOSE 8080
EXPOSE 4317
EXPOSE 4318

ENV OTEL_EXPORTER_OTLP_ENDPOINT=0.0.0.0:4317
ENV OTEL_EXPORTER_OTLP_HTTP_ENDPOINT=0.0.0.0:4318

ENTRYPOINT ["./dice-app"] 