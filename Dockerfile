FROM golang:1.23-alpine

WORKDIR /app

# Copy all source files including internal packages
COPY . .

# Build the runner binary
RUN go build -o runner .

# Pre-warm the Go build cache by compiling a representative program.
# Subsequent user builds reuse this cache, dropping compile time from 30s+ to ~1-2s.
ENV GOCACHE=/root/.cache/go-build
RUN mkdir -p /tmp/warmup && \
    printf 'package main\n\nimport (\n\t"fmt"\n\t"os"\n\t"bufio"\n\t"strings"\n\t"strconv"\n\t"sort"\n\t"math"\n)\n\nfunc main() {\n\t_ = fmt.Sprintf\n\t_ = os.Args\n\t_ = bufio.NewReader\n\t_ = strings.TrimSpace\n\t_ = strconv.Itoa\n\t_ = sort.Ints\n\t_ = math.Abs\n}\n' > /tmp/warmup/main.go && \
    go build -o /tmp/warmup/out /tmp/warmup/main.go && \
    rm -rf /tmp/warmup

EXPOSE 3001

CMD ["./runner"]
