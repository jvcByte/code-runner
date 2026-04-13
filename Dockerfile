FROM golang:1.23-alpine

WORKDIR /app

# Copy all source files including internal packages
COPY . .

# Build the runner binary
RUN go build -o runner .

EXPOSE 3001

CMD ["./runner"]
