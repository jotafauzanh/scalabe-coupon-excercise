# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install git for fetching dependencies
RUN apk add --no-cache git

COPY go.mod ./
# Copy go.sum if it exists, otherwise we will generate it
# COPY go.sum ./ 
# In this environment, go.sum checks might fail if we don't have it yet, 
# so we'll run tidy.

COPY . .

# We run mod tidy here because we just added postgres dependency in the code
# but haven't updated go.mod in the host environment yet.
RUN go mod tidy

RUN go build -o main cmd/server/main.go

# Run stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .

EXPOSE 8080
CMD ["./main"]
