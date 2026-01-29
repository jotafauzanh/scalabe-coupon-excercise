FROM golang:1.23-alpine

WORKDIR /app

# Install git for fetching dependencies
RUN apk add --no-cache git

COPY go.mod ./
COPY go.sum ./

COPY . .

# Make sure deps are ready
RUN go mod tidy

EXPOSE 8080

CMD ["go", "run", "cmd/server/main.go"]