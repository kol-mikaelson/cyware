FROM golang:1.24-alpine
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod ./
RUN go mod download
COPY . .
EXPOSE 8080
CMD ["go", "run", "./cmd/api/main.go"]
