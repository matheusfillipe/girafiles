FROM golang:1.24 as builder
WORKDIR /app

# Copy dependencies first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o build/girafiles .

ARG TESTING
RUN if [ "$TESTING" = "true" ]; then \
    apt-get update && apt-get install -y sqlite3; \
  fi

# Use distroless for smaller attack surface
FROM gcr.io/distroless/base-debian12

WORKDIR /app
COPY --from=builder /app/build/girafiles ./girafiles
COPY --from=builder /app/web ./web

ENV PORT=8080
EXPOSE 8080

CMD ["./girafiles"]
