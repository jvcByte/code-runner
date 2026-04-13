FROM golang:1.22-alpine

# Install git (needed for go run with modules)
RUN apk add --no-cache git

WORKDIR /app

# Copy and build the runner server
COPY main.go .
RUN go mod init runner && go build -o runner .

# The runner binary will use the system Go to execute user code
EXPOSE 3001

CMD ["./runner"]
