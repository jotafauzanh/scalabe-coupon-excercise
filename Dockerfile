# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install git for fetching dependencies
RUN apk add --no-cache git

COPY go.mod ./

# Copy go.sum if it exists, otherwise we will generate it
# COPY go.sum ./

COPY . .

# Make sure deps are ready
RUN go mod tidy

RUN go build -o main cmd/server/main.go

# Run stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .

EXPOSE 8080
CMD ["./main"]
