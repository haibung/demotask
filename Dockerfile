# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app

# Install git for go modules (if needed)
RUN apk add --no-cache git ca-certificates tzdata

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o /bin/app .

# Runtime stage
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app

COPY --from=builder /bin/app /app/app
COPY --from=builder /app/config.yml /app/config.yml
COPY --from=builder /app/static /app/static
COPY --from=builder /app/sql /app/sql

# Optional: adjust to your app port
EXPOSE 8090

USER nonroot:nonroot
ENTRYPOINT ["/app/app"]
CMD ["start"]
