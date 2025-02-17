FROM golang:1.24 as builder
WORKDIR /app
COPY . .
RUN go mod download

RUN CGO_ENABLED=1 GOOS=linux go build -o build/girafiles .
ARG TESTING
RUN \
  if [ "$TESTING" = "true" ]; then \
    apt-get update && apt-get install -y sqlite3; \
  fi

ENV PORT=8080
EXPOSE 8080

CMD ["./build/girafiles"]
